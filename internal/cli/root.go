// Package cli wires Cobra commands for the hitmaker executable.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/zeb-link/hitmaker/v2/internal/config"
	"github.com/zeb-link/hitmaker/v2/internal/simulator"
	"github.com/zeb-link/hitmaker/v2/internal/tui"
)

type rootOptions struct {
	JSON    bool
	Config  bool
	NoIntro bool
	Version string
}

func Execute(version string) {
	opts := &rootOptions{Version: version}
	cmd := newRootCommand(opts)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hitmaker [url...|file]",
		Short: "Synthetic traffic with style",
		Long: strings.TrimSpace(`Hitmaker fires realistic-looking HTTP requests at target URLs with varied identity, schedule, and origin controls.

Interactive traffic (live dashboard — watch hits, press C to configure without stopping):
  hitmaker https://example.com/a
  hitmaker tui links.txt

Headless traffic:
  hitmaker run --for 10m --rate 5-20 https://example.com/a

Bot & AI-crawler traffic (real public User-Agent strings analytics classify as bots):
  hitmaker bots                                      # list every bot/AI identity
  hitmaker run --bots ai --bot-ratio 100 URL         # only AI crawlers + assistants
  hitmaker run --bots GPTBot,ClaudeBot URL           # specific bots by name
  hitmaker probe --bots PerplexityBot URL            # one diagnostic bot hit

Configuration:
  hitmaker config edit
  hitmaker --config`),
		Example: strings.TrimSpace(`hitmaker https://example.com/a
hitmaker tui links.txt
hitmaker run --for 30s --rate 60 --mode vercel https://example.com/a
hitmaker run --bots ai --bot-ratio 100 https://example.com/a
hitmaker bots --json
hitmaker config edit`),
		// Accept bare targets (URLs / files) on the root command. Without this,
		// Cobra's default validator treats the first arg as a subcommand and
		// rejects `hitmaker <url>` with "unknown command". Registered subcommands
		// (run, tui, ...) still match first and take precedence.
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Config {
				return runConfigEditor()
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			return runTrafficTUI(args, opts.NoIntro)
		},
	}
	cmd.PersistentFlags().BoolVarP(&opts.JSON, "json", "j", false, "write machine-readable JSON")
	cmd.Flags().BoolVar(&opts.Config, "config", false, "open the interactive config editor")
	cmd.PersistentFlags().BoolVar(&opts.NoIntro, "no-intro", false, "skip the intro animation")
	cmd.AddCommand(newTUICommand(opts), newRunCommand(opts), newProbeCommand(opts), newBotsCommand(opts), newConfigCommand(opts), newVersionCommand(opts), newFrameCommand(opts))
	cmd.Version = opts.Version
	return cmd
}

func newTUICommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "tui [url...|file]",
		Aliases: []string{"dash", "dashboard"},
		Short:   "Open the interactive traffic dashboard",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrafficTUI(args, opts.NoIntro)
		},
	}
}

func runTrafficTUI(args []string, noIntro bool) error {
	targets, err := expandTargets(args)
	if err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	model, err := tui.New(tui.Options{Config: cfg, Targets: targets, NoIntro: noIntro})
	if err != nil {
		return err
	}
	// Alt-screen is requested per-frame via View.AltScreen in v2.
	_, err = tea.NewProgram(model).Run()
	return err
}

func runConfigEditor() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(tui.NewConfigModel(cfg)).Run()
	return err
}

func newVersionCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.JSON {
				return writeJSON(map[string]string{"version": opts.Version})
			}
			fmt.Printf("hitmaker version %s\n", opts.Version)
			return nil
		},
	}
}

func expandTargets(args []string) ([]string, error) {
	out := []string{}
	for _, arg := range args {
		if looksLikeURL(arg) {
			out = append(out, arg)
			continue
		}
		info, err := os.Stat(arg)
		if err == nil && !info.IsDir() {
			targets, err := readTargetsFile(arg)
			if err != nil {
				return nil, err
			}
			out = append(out, targets...)
			continue
		}
		if err == nil && info.IsDir() {
			return nil, fmt.Errorf("%s is a directory, expected URL or text file", arg)
		}
		if strings.Contains(arg, string(filepath.Separator)) || strings.HasSuffix(arg, ".txt") {
			return nil, fmt.Errorf("read targets file %s: %w", arg, err)
		}
		return nil, fmt.Errorf("target must be an http(s) URL or readable file: %s", arg)
	}
	return simulator.NormalizeTargets(out)
}

func readTargetsFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	targets := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		targets = append(targets, line)
	}
	return targets, scanner.Err()
}

func looksLikeURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func writeJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

// writeJSONLine emits one compact JSON object per line (NDJSON), which is what
// a streaming consumer wants for periodic snapshots.
func writeJSONLine(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(append(data, '\n'))
	return err
}

func parseRate(value string, cfg *config.Config) error {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, "-")
	if len(parts) == 1 {
		var rate int
		if _, err := fmt.Sscanf(parts[0], "%d", &rate); err != nil {
			return fmt.Errorf("--rate must be N or MIN-MAX")
		}
		cfg.Traffic.MinPerMin = rate
		cfg.Traffic.MaxPerMin = rate
		return nil
	}
	if len(parts) == 2 {
		var min, max int
		if _, err := fmt.Sscanf(parts[0], "%d", &min); err != nil {
			return fmt.Errorf("--rate must be N or MIN-MAX")
		}
		if _, err := fmt.Sscanf(parts[1], "%d", &max); err != nil {
			return fmt.Errorf("--rate must be N or MIN-MAX")
		}
		cfg.Traffic.MinPerMin = min
		cfg.Traffic.MaxPerMin = max
		return nil
	}
	return fmt.Errorf("--rate must be N or MIN-MAX")
}

func durationOrZero(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	return time.ParseDuration(value)
}
