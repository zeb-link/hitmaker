package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zeb-link/hitmaker/v2/internal/config"
)

func newConfigCommand(root *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and edit configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			return writeJSON(cfg.Redacted())
		},
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:     "edit",
			Aliases: []string{"tui"},
			Short:   "Open the interactive config editor",
			RunE: func(cmd *cobra.Command, args []string) error {
				return runConfigEditor()
			},
		},
		&cobra.Command{
			Use:   "print",
			Short: "Print resolved configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return err
				}
				return writeJSON(cfg.Redacted())
			},
		},
		&cobra.Command{
			Use:   "set <key> <value>",
			Short: "Set a global config value",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return err
				}
				if err := setKey(&cfg, args[0], args[1]); err != nil {
					return err
				}
				return config.SaveGlobal(cfg)
			},
		},
		&cobra.Command{
			Use:   "reset",
			Short: "Remove global config",
			RunE: func(cmd *cobra.Command, args []string) error {
				return config.ResetGlobal()
			},
		},
	)
	return cmd
}

func setKey(cfg *config.Config, key, raw string) error {
	switch strings.ToLower(key) {
	case "traffic.minpermin", "minpermin", "min_per_min":
		value, err := strconv.Atoi(raw)
		cfg.Traffic.MinPerMin = value
		return err
	case "traffic.maxpermin", "maxpermin", "max_per_min":
		value, err := strconv.Atoi(raw)
		cfg.Traffic.MaxPerMin = value
		return err
	case "traffic.concurrent", "concurrent":
		value, err := strconv.Atoi(raw)
		cfg.Traffic.Concurrent = value
		return err
	case "traffic.timeoutms", "timeoutms", "timeout_ms":
		value, err := strconv.Atoi(raw)
		cfg.Traffic.TimeoutMs = value
		return err
	case "requests.method", "method":
		cfg.Requests.Method = strings.ToUpper(raw)
	case "requests.deviceratio", "deviceratio", "device_ratio":
		value, err := strconv.Atoi(raw)
		cfg.Requests.DeviceRatio = value
		return err
	case "requests.unknownratio", "unknownratio", "unknown_ratio":
		value, err := strconv.Atoi(raw)
		cfg.Requests.UnknownRatio = value
		return err
	case "requests.uniqueipprob", "uniqueipprob", "unique_ip_prob":
		value, err := strconv.ParseFloat(raw, 64)
		cfg.Requests.UniqueIPProb = value
		return err
	case "requests.bots", "bots":
		if strings.TrimSpace(raw) == "" || strings.EqualFold(raw, "all") || strings.EqualFold(raw, "none") {
			cfg.Requests.Bots = nil
		} else {
			cfg.Requests.Bots = splitBotTokens(raw)
		}
	case "schedule.idleodds", "idleodds", "idle_odds":
		value, err := strconv.ParseFloat(raw, 64)
		cfg.Schedule.IdleOdds = value
		return err
	case "entropy.level", "entropy":
		cfg.Entropy.Level = config.EntropyLevel(strings.ToLower(strings.TrimSpace(raw)))
	case "entropy.devicespread", "devicespread", "device_spread":
		value, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		cfg.Entropy.DeviceSpread = value
		cfg.Entropy.Level = config.EntropyCustom
	case "entropy.breakout", "breakout":
		value, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		cfg.Entropy.Breakout = value
		cfg.Entropy.Level = config.EntropyCustom
	case "entropy.viralpercent", "viralpercent", "viral_percent":
		value, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		cfg.Entropy.ViralPercent = value
		cfg.Entropy.Level = config.EntropyCustom
	case "origin.mode", "mode":
		cfg.Origin.Mode = config.Mode(raw)
	case "origin.provider", "provider":
		cfg.Origin.Provider = raw
	case "origin.iproyalurl", "iproyalurl", "iproyal_url":
		if cfg.Origin.ProviderConfig == nil {
			cfg.Origin.ProviderConfig = map[string]string{}
		}
		cfg.Origin.Provider = "iproyal"
		cfg.Origin.ProviderConfig["url"] = raw
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return cfg.Validate()
}
