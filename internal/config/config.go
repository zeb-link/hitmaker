// Package config owns Hitmaker's typed runtime configuration and persistence.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zeb-link/hitmaker/internal/identity"
)

const (
	LocalConfigFile = ".hitmaker.json"
)

type Mode string

const (
	ModeNone   Mode = "none"
	ModeAuto   Mode = "auto"
	ModeVercel Mode = "vercel"
	ModeProxy  Mode = "proxy"
)

type Config struct {
	Traffic  TrafficConfig  `json:"traffic"`
	Requests RequestConfig  `json:"requests"`
	Schedule ScheduleConfig `json:"schedule"`
	Origin   OriginConfig   `json:"origin"`
}

type TrafficConfig struct {
	MinPerMin  int `json:"minPerMin"`
	MaxPerMin  int `json:"maxPerMin"`
	Concurrent int `json:"concurrent"`
	TimeoutMs  int `json:"timeoutMs"`
}

type RequestConfig struct {
	Method       string     `json:"method"`
	DeviceRatio  int        `json:"deviceRatio"`
	UnknownRatio int        `json:"unknownRatio"`
	UniqueIPProb float64    `json:"uniqueIpProb"`
	URLParams    []URLParam `json:"urlParams"`
	// Bots restricts the unknown/bot pool to specific categories (e.g.
	// "ai_crawler") or exact names (e.g. "GPTBot"). Empty means the whole
	// catalog. UnknownRatio still governs how much traffic is bots.
	Bots []string `json:"bots,omitempty"`
}

type ScheduleConfig struct {
	MinActive int     `json:"minActive"`
	MaxActive int     `json:"maxActive"`
	IdleOdds  float64 `json:"idleOdds"`
	MinIdle   int     `json:"minIdle"`
	MaxIdle   int     `json:"maxIdle"`
}

type OriginConfig struct {
	Mode           Mode              `json:"mode"`
	Provider       string            `json:"provider,omitempty"`
	ProviderConfig map[string]string `json:"providerConfig,omitempty"`
}

type URLParam struct {
	Key         string    `json:"key"`
	Value       string    `json:"value,omitempty"`
	Probability float64   `json:"probability"`
	Payloads    []Payload `json:"payloads,omitempty"`
}

type Payload struct {
	Name   string            `json:"name"`
	Weight float64           `json:"weight"`
	KV     map[string]string `json:"params"`
}

type Partial struct {
	Traffic  *PartialTraffic  `json:"traffic,omitempty"`
	Requests *PartialRequest  `json:"requests,omitempty"`
	Schedule *PartialSchedule `json:"schedule,omitempty"`
	Origin   *PartialOrigin   `json:"origin,omitempty"`
}

type PartialTraffic struct {
	MinPerMin  *int `json:"minPerMin,omitempty"`
	MaxPerMin  *int `json:"maxPerMin,omitempty"`
	Concurrent *int `json:"concurrent,omitempty"`
	TimeoutMs  *int `json:"timeoutMs,omitempty"`
}

type PartialRequest struct {
	Method       *string     `json:"method,omitempty"`
	DeviceRatio  *int        `json:"deviceRatio,omitempty"`
	UnknownRatio *int        `json:"unknownRatio,omitempty"`
	UniqueIPProb *float64    `json:"uniqueIpProb,omitempty"`
	URLParams    *[]URLParam `json:"urlParams,omitempty"`
	Bots         *[]string   `json:"bots,omitempty"`
}

type PartialSchedule struct {
	MinActive *int     `json:"minActive,omitempty"`
	MaxActive *int     `json:"maxActive,omitempty"`
	IdleOdds  *float64 `json:"idleOdds,omitempty"`
	MinIdle   *int     `json:"minIdle,omitempty"`
	MaxIdle   *int     `json:"maxIdle,omitempty"`
}

type PartialOrigin struct {
	Mode           *Mode              `json:"mode,omitempty"`
	Provider       *string            `json:"provider,omitempty"`
	ProviderConfig *map[string]string `json:"providerConfig,omitempty"`
}

func Default() Config {
	return Config{
		Traffic: TrafficConfig{
			MinPerMin:  1,
			MaxPerMin:  25,
			Concurrent: 1,
			TimeoutMs:  5000,
		},
		Requests: RequestConfig{
			Method:       "GET",
			DeviceRatio:  60,
			UnknownRatio: 5,
			UniqueIPProb: 0.95,
			// No default URL params: targets are opaque URLs, and an
			// injected param silently pollutes whatever query-string
			// semantics the target system has. Params are strictly
			// opt-in via config.
			URLParams: nil,
		},
		Schedule: ScheduleConfig{
			MinActive: 5,
			MaxActive: 15,
			IdleOdds:  0.75,
			MinIdle:   1,
			MaxIdle:   15,
		},
		Origin: OriginConfig{
			Mode: ModeNone,
		},
	}
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hitmaker"), nil
}

func GlobalPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LocalPath() string {
	return filepath.Join(".", LocalConfigFile)
}

func Load() (Config, error) {
	cfg := Default()
	if partial, ok, err := readPartialGlobal(); err != nil {
		return Config{}, err
	} else if ok {
		merge(&cfg, partial)
	}
	if partial, ok, err := readPartial(LocalPath()); err != nil {
		return Config{}, err
	} else if ok {
		merge(&cfg, partial)
	}
	env, err := envPartial()
	if err != nil {
		return Config{}, err
	}
	merge(&cfg, env)
	return cfg, cfg.Validate()
}

func SaveGlobal(cfg Config) error {
	path, err := GlobalPath()
	if err != nil {
		return err
	}
	return writeJSON(path, cfg, true)
}

func SaveLocal(cfg Config) error {
	return writeJSON(LocalPath(), cfg, false)
}

func ResetGlobal() error {
	path, err := GlobalPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (c Config) Validate() error {
	if c.Traffic.MinPerMin < 0 || c.Traffic.MaxPerMin < 0 {
		return errors.New("traffic rates must be >= 0")
	}
	if c.Traffic.MaxPerMin < c.Traffic.MinPerMin {
		return errors.New("maxPerMin must be >= minPerMin")
	}
	if c.Traffic.Concurrent < 1 {
		return errors.New("concurrent must be >= 1")
	}
	if c.Traffic.TimeoutMs < 100 {
		return errors.New("timeoutMs must be >= 100")
	}
	method := strings.ToUpper(c.Requests.Method)
	if method != "GET" && method != "HEAD" && method != "POST" {
		return errors.New("method must be GET, HEAD, or POST")
	}
	if c.Requests.DeviceRatio < 0 || c.Requests.DeviceRatio > 100 {
		return errors.New("deviceRatio must be 0..100")
	}
	if c.Requests.UnknownRatio < 0 || c.Requests.UnknownRatio > 100 {
		return errors.New("unknownRatio must be 0..100")
	}
	if c.Requests.UniqueIPProb < 0 || c.Requests.UniqueIPProb > 1 {
		return errors.New("uniqueIpProb must be 0..1")
	}
	if c.Schedule.MinActive < 0 || c.Schedule.MaxActive < 0 || c.Schedule.MaxActive < c.Schedule.MinActive {
		return errors.New("active phase minutes are invalid")
	}
	if c.Schedule.IdleOdds < 0 || c.Schedule.IdleOdds > 1 {
		return errors.New("idleOdds must be 0..1")
	}
	if c.Schedule.MinIdle < 0 || c.Schedule.MaxIdle < 0 || c.Schedule.MaxIdle < c.Schedule.MinIdle {
		return errors.New("idle phase minutes are invalid")
	}
	switch c.Origin.Mode {
	case ModeNone, ModeAuto, ModeVercel, ModeProxy:
	default:
		return fmt.Errorf("unknown origin mode %q", c.Origin.Mode)
	}
	if _, err := c.BotFilter(); err != nil {
		return err
	}
	return nil
}

// BotFilter resolves the configured bot spec into an identity.BotFilter. An
// empty spec yields an empty (match-all) filter.
func (c Config) BotFilter() (identity.BotFilter, error) {
	return identity.ParseBotSpec(c.Requests.Bots)
}

func splitTokens(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if field = strings.TrimSpace(field); field != "" {
			out = append(out, field)
		}
	}
	return out
}

func (c Config) Redacted() Config {
	cp := c
	if len(cp.Origin.ProviderConfig) > 0 {
		cp.Origin.ProviderConfig = map[string]string{}
		for key, value := range c.Origin.ProviderConfig {
			if value == "" {
				cp.Origin.ProviderConfig[key] = ""
			} else {
				cp.Origin.ProviderConfig[key] = "********"
			}
		}
	}
	return cp
}

func readPartialGlobal() (Partial, bool, error) {
	path, err := GlobalPath()
	if err != nil {
		return Partial{}, false, err
	}
	return readPartial(path)
}

func readPartial(path string) (Partial, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Partial{}, false, nil
	}
	if err != nil {
		return Partial{}, false, err
	}
	var partial Partial
	if err := json.Unmarshal(data, &partial); err != nil {
		return Partial{}, false, fmt.Errorf("load %s: %w", path, err)
	}
	legacy, err := legacyPartial(data)
	if err != nil {
		return Partial{}, false, fmt.Errorf("load %s: %w", path, err)
	}
	mergePartial(&partial, legacy)
	return partial, true, nil
}

func writeJSON(path string, cfg Config, makeParent bool) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if makeParent {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func merge(cfg *Config, partial Partial) {
	if partial.Traffic != nil {
		if partial.Traffic.MinPerMin != nil {
			cfg.Traffic.MinPerMin = *partial.Traffic.MinPerMin
		}
		if partial.Traffic.MaxPerMin != nil {
			cfg.Traffic.MaxPerMin = *partial.Traffic.MaxPerMin
		}
		if partial.Traffic.Concurrent != nil {
			cfg.Traffic.Concurrent = *partial.Traffic.Concurrent
		}
		if partial.Traffic.TimeoutMs != nil {
			cfg.Traffic.TimeoutMs = *partial.Traffic.TimeoutMs
		}
	}
	if partial.Requests != nil {
		if partial.Requests.Method != nil {
			cfg.Requests.Method = strings.ToUpper(*partial.Requests.Method)
		}
		if partial.Requests.DeviceRatio != nil {
			cfg.Requests.DeviceRatio = *partial.Requests.DeviceRatio
		}
		if partial.Requests.UnknownRatio != nil {
			cfg.Requests.UnknownRatio = *partial.Requests.UnknownRatio
		}
		if partial.Requests.UniqueIPProb != nil {
			cfg.Requests.UniqueIPProb = *partial.Requests.UniqueIPProb
		}
		if partial.Requests.URLParams != nil {
			cfg.Requests.URLParams = *partial.Requests.URLParams
		}
		if partial.Requests.Bots != nil {
			cfg.Requests.Bots = *partial.Requests.Bots
		}
	}
	if partial.Schedule != nil {
		if partial.Schedule.MinActive != nil {
			cfg.Schedule.MinActive = *partial.Schedule.MinActive
		}
		if partial.Schedule.MaxActive != nil {
			cfg.Schedule.MaxActive = *partial.Schedule.MaxActive
		}
		if partial.Schedule.IdleOdds != nil {
			cfg.Schedule.IdleOdds = *partial.Schedule.IdleOdds
		}
		if partial.Schedule.MinIdle != nil {
			cfg.Schedule.MinIdle = *partial.Schedule.MinIdle
		}
		if partial.Schedule.MaxIdle != nil {
			cfg.Schedule.MaxIdle = *partial.Schedule.MaxIdle
		}
	}
	if partial.Origin != nil {
		if partial.Origin.Mode != nil {
			cfg.Origin.Mode = *partial.Origin.Mode
		}
		if partial.Origin.Provider != nil {
			cfg.Origin.Provider = *partial.Origin.Provider
		}
		if partial.Origin.ProviderConfig != nil {
			cfg.Origin.ProviderConfig = *partial.Origin.ProviderConfig
		}
	}
}

func mergePartial(dst *Partial, src Partial) {
	if src.Traffic != nil {
		if dst.Traffic == nil {
			dst.Traffic = &PartialTraffic{}
		}
		if src.Traffic.MinPerMin != nil {
			dst.Traffic.MinPerMin = src.Traffic.MinPerMin
		}
		if src.Traffic.MaxPerMin != nil {
			dst.Traffic.MaxPerMin = src.Traffic.MaxPerMin
		}
		if src.Traffic.Concurrent != nil {
			dst.Traffic.Concurrent = src.Traffic.Concurrent
		}
		if src.Traffic.TimeoutMs != nil {
			dst.Traffic.TimeoutMs = src.Traffic.TimeoutMs
		}
	}
	if src.Requests != nil {
		if dst.Requests == nil {
			dst.Requests = &PartialRequest{}
		}
		if src.Requests.Method != nil {
			dst.Requests.Method = src.Requests.Method
		}
		if src.Requests.DeviceRatio != nil {
			dst.Requests.DeviceRatio = src.Requests.DeviceRatio
		}
		if src.Requests.UnknownRatio != nil {
			dst.Requests.UnknownRatio = src.Requests.UnknownRatio
		}
		if src.Requests.UniqueIPProb != nil {
			dst.Requests.UniqueIPProb = src.Requests.UniqueIPProb
		}
		if src.Requests.URLParams != nil {
			dst.Requests.URLParams = src.Requests.URLParams
		}
		if src.Requests.Bots != nil {
			dst.Requests.Bots = src.Requests.Bots
		}
	}
	if src.Schedule != nil {
		if dst.Schedule == nil {
			dst.Schedule = &PartialSchedule{}
		}
		if src.Schedule.MinActive != nil {
			dst.Schedule.MinActive = src.Schedule.MinActive
		}
		if src.Schedule.MaxActive != nil {
			dst.Schedule.MaxActive = src.Schedule.MaxActive
		}
		if src.Schedule.IdleOdds != nil {
			dst.Schedule.IdleOdds = src.Schedule.IdleOdds
		}
		if src.Schedule.MinIdle != nil {
			dst.Schedule.MinIdle = src.Schedule.MinIdle
		}
		if src.Schedule.MaxIdle != nil {
			dst.Schedule.MaxIdle = src.Schedule.MaxIdle
		}
	}
	if src.Origin != nil {
		if dst.Origin == nil {
			dst.Origin = &PartialOrigin{}
		}
		if src.Origin.Mode != nil {
			dst.Origin.Mode = src.Origin.Mode
		}
		if src.Origin.Provider != nil {
			dst.Origin.Provider = src.Origin.Provider
		}
		if src.Origin.ProviderConfig != nil {
			dst.Origin.ProviderConfig = src.Origin.ProviderConfig
		}
	}
}

type legacyFile struct {
	MinPerMin       *int        `json:"MIN_PER_MIN"`
	MaxPerMin       *int        `json:"MAX_PER_MIN"`
	Concurrent      *int        `json:"CONCURRENT"`
	Method          *string     `json:"METHOD"`
	TimeoutMs       *int        `json:"TIMEOUT_MS"`
	DeviceRatio     *int        `json:"DEVICE_RATIO"`
	UnknownRatio    *int        `json:"UNKNOWN_RATIO"`
	MinActive       *int        `json:"MIN_ACTIVE"`
	MaxActive       *int        `json:"MAX_ACTIVE"`
	IdleOdds        *float64    `json:"IDLE_ODDS"`
	MinIdle         *int        `json:"MIN_IDLE"`
	MaxIdle         *int        `json:"MAX_IDLE"`
	UniqueIPProb    *float64    `json:"UNIQUE_IP_PROB"`
	ProxyMode       *string     `json:"PROXY_MODE"`
	ProxyServiceURL *string     `json:"PROXY_SERVICE_URL"`
	URLParams       *[]URLParam `json:"URL_PARAMS"`
}

func legacyPartial(data []byte) (Partial, error) {
	var legacy legacyFile
	if err := json.Unmarshal(data, &legacy); err != nil {
		return Partial{}, err
	}
	var p Partial
	if legacy.MinPerMin != nil || legacy.MaxPerMin != nil || legacy.Concurrent != nil || legacy.TimeoutMs != nil {
		p.Traffic = &PartialTraffic{
			MinPerMin:  legacy.MinPerMin,
			MaxPerMin:  legacy.MaxPerMin,
			Concurrent: legacy.Concurrent,
			TimeoutMs:  legacy.TimeoutMs,
		}
	}
	if legacy.Method != nil || legacy.DeviceRatio != nil || legacy.UnknownRatio != nil || legacy.UniqueIPProb != nil || legacy.URLParams != nil {
		p.Requests = &PartialRequest{
			Method:       legacy.Method,
			DeviceRatio:  legacy.DeviceRatio,
			UnknownRatio: legacy.UnknownRatio,
			UniqueIPProb: legacy.UniqueIPProb,
			URLParams:    legacy.URLParams,
		}
	}
	if legacy.MinActive != nil || legacy.MaxActive != nil || legacy.IdleOdds != nil || legacy.MinIdle != nil || legacy.MaxIdle != nil {
		p.Schedule = &PartialSchedule{
			MinActive: legacy.MinActive,
			MaxActive: legacy.MaxActive,
			IdleOdds:  legacy.IdleOdds,
			MinIdle:   legacy.MinIdle,
			MaxIdle:   legacy.MaxIdle,
		}
	}
	if legacy.ProxyMode != nil || legacy.ProxyServiceURL != nil {
		p.Origin = &PartialOrigin{}
		if legacy.ProxyMode != nil {
			mode := legacyMode(*legacy.ProxyMode)
			p.Origin.Mode = &mode
		}
		if legacy.ProxyServiceURL != nil && *legacy.ProxyServiceURL != "" {
			provider := "iproyal"
			providerCfg := map[string]string{"url": *legacy.ProxyServiceURL}
			mode := ModeProxy
			p.Origin.Mode = &mode
			p.Origin.Provider = &provider
			p.Origin.ProviderConfig = &providerCfg
		}
	}
	return p, nil
}

func envPartial() (Partial, error) {
	var p Partial
	t := &PartialTraffic{}
	r := &PartialRequest{}
	s := &PartialSchedule{}
	o := &PartialOrigin{}

	var usedT, usedR, usedS, usedO bool
	if value, ok, err := envInt("MIN_PER_MIN"); err != nil {
		return p, err
	} else if ok {
		t.MinPerMin = &value
		usedT = true
	}
	if value, ok, err := envInt("MAX_PER_MIN"); err != nil {
		return p, err
	} else if ok {
		t.MaxPerMin = &value
		usedT = true
	}
	if value, ok, err := envInt("CONCURRENT"); err != nil {
		return p, err
	} else if ok {
		t.Concurrent = &value
		usedT = true
	}
	if value, ok, err := envInt("TIMEOUT_MS"); err != nil {
		return p, err
	} else if ok {
		t.TimeoutMs = &value
		usedT = true
	}
	if value, ok := os.LookupEnv("METHOD"); ok {
		r.Method = &value
		usedR = true
	}
	if value, ok, err := envInt("DEVICE_RATIO"); err != nil {
		return p, err
	} else if ok {
		r.DeviceRatio = &value
		usedR = true
	}
	if value, ok, err := envInt("UNKNOWN_RATIO"); err != nil {
		return p, err
	} else if ok {
		r.UnknownRatio = &value
		usedR = true
	}
	if value, ok, err := envFloat("UNIQUE_IP_PROB"); err != nil {
		return p, err
	} else if ok {
		r.UniqueIPProb = &value
		usedR = true
	}
	if raw, ok := os.LookupEnv("URL_PARAMS"); ok {
		var params []URLParam
		if err := json.Unmarshal([]byte(raw), &params); err != nil {
			return p, fmt.Errorf("URL_PARAMS: %w", err)
		}
		r.URLParams = &params
		usedR = true
	}
	if raw, ok := os.LookupEnv("BOTS"); ok {
		tokens := splitTokens(raw)
		r.Bots = &tokens
		usedR = true
	}
	if value, ok, err := envInt("MIN_ACTIVE"); err != nil {
		return p, err
	} else if ok {
		s.MinActive = &value
		usedS = true
	}
	if value, ok, err := envInt("MAX_ACTIVE"); err != nil {
		return p, err
	} else if ok {
		s.MaxActive = &value
		usedS = true
	}
	if value, ok, err := envFloat("IDLE_ODDS"); err != nil {
		return p, err
	} else if ok {
		s.IdleOdds = &value
		usedS = true
	}
	if value, ok, err := envInt("MIN_IDLE"); err != nil {
		return p, err
	} else if ok {
		s.MinIdle = &value
		usedS = true
	}
	if value, ok, err := envInt("MAX_IDLE"); err != nil {
		return p, err
	} else if ok {
		s.MaxIdle = &value
		usedS = true
	}
	if value, ok := os.LookupEnv("HITMAKER_MODE"); ok {
		mode := Mode(value)
		o.Mode = &mode
		usedO = true
	} else if value, ok := os.LookupEnv("PROXY_MODE"); ok {
		mode := legacyMode(value)
		o.Mode = &mode
		usedO = true
	}
	if value, ok := os.LookupEnv("HITMAKER_PROVIDER"); ok {
		o.Provider = &value
		usedO = true
	}
	if value, ok := os.LookupEnv("IPROYAL_URL"); ok {
		cfg := map[string]string{"url": value}
		o.ProviderConfig = &cfg
		provider := "iproyal"
		o.Provider = &provider
		mode := ModeProxy
		o.Mode = &mode
		usedO = true
	} else if value, ok := os.LookupEnv("PROXY_SERVICE_URL"); ok {
		cfg := map[string]string{"url": value}
		o.ProviderConfig = &cfg
		provider := "iproyal"
		o.Provider = &provider
		mode := ModeProxy
		o.Mode = &mode
		usedO = true
	}

	if usedT {
		p.Traffic = t
	}
	if usedR {
		p.Requests = r
	}
	if usedS {
		p.Schedule = s
	}
	if usedO {
		p.Origin = o
	}
	return p, nil
}

func envInt(key string) (int, bool, error) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return 0, false, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, true, fmt.Errorf("%s: %w", key, err)
	}
	return value, true, nil
}

func envFloat(key string) (float64, bool, error) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return 0, false, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, true, fmt.Errorf("%s: %w", key, err)
	}
	return value, true, nil
}

func legacyMode(value string) Mode {
	switch strings.ToLower(value) {
	case "auto":
		return ModeAuto
	case "service", "url", "free", "proxy":
		return ModeProxy
	case "vercel":
		return ModeVercel
	default:
		return ModeNone
	}
}
