package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/zeb-link/hitmaker/internal/config"
	"github.com/zeb-link/hitmaker/internal/simulator"
	"github.com/zeb-link/hitmaker/internal/ui/theme"
)

type probeOptions struct {
	Mode     string
	Factory  bool
	Timeout  time.Duration
	Bots     string
	BotRatio int
	Follow   bool
	Seed     int64
}

func newProbeCommand(root *rootOptions) *cobra.Command {
	opts := &probeOptions{Timeout: 10 * time.Second}
	cmd := &cobra.Command{
		Use:   "probe <url>",
		Short: "Send one diagnostic request",
		Long:  "Send one request using Hitmaker's identity/origin stack and print the result. This is the fastest way to test a target.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigForRun(opts.Factory)
			if err != nil {
				return err
			}
			if opts.Mode != "" {
				cfg.Origin.Mode = config.Mode(opts.Mode)
			}
			if err := applyBotFlags(&cfg, opts.Bots, opts.BotRatio, cmd.Flags().Changed("bot-ratio")); err != nil {
				return err
			}
			cfg.Traffic.TimeoutMs = int(opts.Timeout / time.Millisecond)
			ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout+time.Second)
			defer cancel()
			runner, err := simulator.New(ctx, simulator.Options{
				Config:  cfg,
				Targets: []string{args[0]},
				Follow:  opts.Follow,
				Seed:    opts.Seed,
			})
			if err != nil {
				return err
			}
			result := runner.Probe(args[0])
			runner.Stop()
			if root.JSON {
				return writeJSON(result)
			}
			printProbeResult(cfg, result)
			if result.Err != "" {
				return fmt.Errorf("probe failed")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "origin mode: none, auto, vercel, proxy")
	cmd.Flags().BoolVar(&opts.Factory, "factory", false, "ignore saved config and use built-in defaults")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 10*time.Second, "request timeout")
	cmd.Flags().StringVar(&opts.Bots, "bots", "", "send a bot identity: category/name, e.g. GPTBot or ai (see `hitmaker bots`)")
	cmd.Flags().IntVar(&opts.BotRatio, "bot-ratio", 0, "percent chance the probe is a bot (0-100)")
	cmd.Flags().BoolVar(&opts.Follow, "follow", false, "follow 3xx redirects (default off: report the redirect's own status)")
	cmd.Flags().Int64Var(&opts.Seed, "seed", 0, "seed for a reproducible identity (0 = random)")
	return cmd
}

func printProbeResult(cfg config.Config, result simulator.HitResult) {
	fmt.Printf("%s %s\n", theme.Logo.Render("HITMAKER probe"), result.Target)
	fmt.Printf("mode=%s timeout=%dms\n", cfg.Origin.Mode, cfg.Traffic.TimeoutMs)
	if result.Err != "" {
		fmt.Printf("%s %s\n", theme.Bad.Render("error"), result.Err)
		return
	}
	fmt.Printf("%s status=%d latency=%s\n", theme.Good.Render("ok"), result.Status, result.Latency.Truncate(time.Millisecond))
	if result.UserAgent != "" {
		fmt.Printf("ua=%s\n", result.UserAgent)
	}
	if result.Location != "" {
		fmt.Printf("identity=%s ip=%s\n", result.Location, result.IP)
	}
	if len(result.Applied) > 0 {
		fmt.Printf("params=%v\n", result.Applied)
	}
}
