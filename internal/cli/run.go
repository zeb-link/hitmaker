package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/kerns/hitmaker/internal/config"
	"github.com/kerns/hitmaker/internal/identity"
	"github.com/kerns/hitmaker/internal/simulator"
	"github.com/kerns/hitmaker/internal/ui/theme"
)

type runOptions struct {
	Targets     []string
	For         string
	Rate        string
	Mode        string
	Concurrent  int
	DeviceRatio int
	Interval    time.Duration
	Factory     bool
	Bots        string
	BotRatio    int
	Follow      bool
	Seed        int64
	Summary     bool
}

func newRunCommand(root *rootOptions) *cobra.Command {
	opts := &runOptions{Interval: time.Second}
	cmd := &cobra.Command{
		Use:   "run [url...|file]",
		Short: "Run headless traffic generation",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetArgs := append([]string{}, args...)
			targetArgs = append(targetArgs, opts.Targets...)
			targets, err := expandTargets(targetArgs)
			if err != nil {
				return err
			}
			cfg, err := loadConfigForRun(opts.Factory)
			if err != nil {
				return err
			}
			if err := parseRate(opts.Rate, &cfg); err != nil {
				return err
			}
			if opts.Concurrent > 0 {
				cfg.Traffic.Concurrent = opts.Concurrent
			}
			if cmd.Flags().Changed("device-ratio") {
				if opts.DeviceRatio < 0 || opts.DeviceRatio > 100 {
					return fmt.Errorf("--device-ratio must be 0-100")
				}
				cfg.Requests.DeviceRatio = opts.DeviceRatio
			}
			if opts.Mode != "" {
				cfg.Origin.Mode = config.Mode(opts.Mode)
			}
			if err := applyBotFlags(&cfg, opts.Bots, opts.BotRatio, cmd.Flags().Changed("bot-ratio")); err != nil {
				return err
			}
			runFor, err := durationOrZero(opts.For)
			if err != nil {
				return err
			}
			return runHeadless(root, cfg, targets, runFor, opts)
		},
	}
	cmd.Flags().StringSliceVarP(&opts.Targets, "targets", "t", nil, "target URLs or files")
	cmd.Flags().StringVar(&opts.For, "for", "", "run duration, e.g. 30s, 10m, 1h (omit to run until Ctrl-C)")
	cmd.Flags().StringVar(&opts.Rate, "rate", "", "hits per minute, per worker, as N or MIN-MAX")
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "origin mode: none, vercel, proxy")
	cmd.Flags().IntVarP(&opts.Concurrent, "concurrent", "c", 0, "workers per target")
	cmd.Flags().IntVar(&opts.DeviceRatio, "device-ratio", 0, "percent of human hits that are desktop vs mobile (0-100)")
	cmd.Flags().DurationVar(&opts.Interval, "interval", time.Second, "stats print interval")
	cmd.Flags().BoolVar(&opts.Factory, "factory", false, "ignore saved config and use built-in defaults")
	cmd.Flags().StringVar(&opts.Bots, "bots", "", "restrict bot pool to categories/names, e.g. ai, crawler, GPTBot,ClaudeBot (see `hitmaker bots`)")
	cmd.Flags().IntVar(&opts.BotRatio, "bot-ratio", 0, "percent of traffic that is bots (0-100); same knob as config unknownRatio")
	cmd.Flags().BoolVar(&opts.Follow, "follow", false, "follow 3xx redirects (default off: report the redirect's own status)")
	cmd.Flags().Int64Var(&opts.Seed, "seed", 0, "seed for reproducible identities/schedule (0 = random)")
	cmd.Flags().BoolVar(&opts.Summary, "summary", false, "suppress per-interval output; print only the final totals on exit")
	return cmd
}

// applyBotFlags folds --bots / --bot-ratio into the config. Passing --bots
// without an explicit --bot-ratio implies 100% bot traffic, since asking for a
// bot category almost always means "send me that kind of traffic".
func applyBotFlags(cfg *config.Config, bots string, botRatio int, ratioSet bool) error {
	if bots != "" {
		tokens := splitBotTokens(bots)
		if _, err := identity.ParseBotSpec(tokens); err != nil {
			return err
		}
		cfg.Requests.Bots = tokens
		if !ratioSet {
			cfg.Requests.UnknownRatio = 100
		}
	}
	if ratioSet {
		if botRatio < 0 || botRatio > 100 {
			return fmt.Errorf("--bot-ratio must be 0-100")
		}
		cfg.Requests.UnknownRatio = botRatio
	}
	return cfg.Validate()
}

func loadConfigForRun(factory bool) (config.Config, error) {
	if factory {
		return config.Default(), nil
	}
	return config.Load()
}

func runHeadless(root *rootOptions, cfg config.Config, targets []string, runFor time.Duration, opts *runOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if runFor > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, runFor)
		defer cancel()
	}
	runner, err := simulator.New(ctx, simulator.Options{
		Config:  cfg,
		Targets: targets,
		Follow:  opts.Follow,
		Seed:    opts.Seed,
	})
	if err != nil {
		return err
	}
	if !root.JSON && !opts.Summary {
		printRunConfigHint(cfg)
	}
	runner.Start()
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()
	frame := 0
	for {
		select {
		case <-ctx.Done():
			// Bounded shutdown so the process can never linger past the deadline
			// even if a worker is stuck in an uncancellable proxied connection.
			runner.StopAndWait(5 * time.Second)
			return printFinal(root, runner.Snapshot())
		case <-ticker.C:
			if opts.Summary {
				continue // only the final totals are printed, on exit
			}
			snap := runner.Snapshot()
			if root.JSON {
				if err := writeJSONLine(snap); err != nil {
					return err
				}
			} else {
				printHumanSnapshot(snap, frame)
				frame++
			}
		}
	}
}

func printRunConfigHint(cfg config.Config) {
	fmt.Printf("%s mode=%s rate=%d-%d/min concurrent=%d timeout=%dms\n",
		theme.Subtle.Render("config"),
		cfg.Origin.Mode,
		cfg.Traffic.MinPerMin,
		cfg.Traffic.MaxPerMin,
		cfg.Traffic.Concurrent,
		cfg.Traffic.TimeoutMs,
	)
	if cfg.Origin.Mode == config.ModeProxy {
		fmt.Printf("%s saved config is using proxy mode; use --mode none or --factory for a direct smoke test\n", theme.Warn.Render("warning"))
	}
}

func printFinal(root *rootOptions, snap simulator.Snapshot) error {
	if root.JSON {
		return writeJSONLine(snap)
	}
	fmt.Println()
	fmt.Println(theme.Logo.Render("HITMAKER summary"))
	for _, target := range snap.Targets {
		fmt.Printf("  %s  hits=%d errors=%d rate=%d/min\n", trimTarget(target.Target, 54), target.Hits, target.Errors, target.CurrentRate)
	}
	return nil
}

func printHumanSnapshot(snap simulator.Snapshot, frame int) {
	loader := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	spin := loader[frame%len(loader)]
	fmt.Printf("\r%s %s hits %d  errors %d  workers %d  uptime %s",
		theme.Focus.Render(spin),
		theme.Logo.Render("hitmaker"),
		snap.TotalHits,
		snap.TotalErrors,
		snap.WorkerCount,
		snap.Uptime.Truncate(time.Second),
	)
	if snap.WorkerCapHit {
		fmt.Printf("  %s", theme.Warn.Render("worker cap active"))
	}
}

func trimTarget(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max < 2 {
		return value[:max]
	}
	return value[:max-1] + "…"
}
