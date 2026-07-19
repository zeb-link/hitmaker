package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeHonorsZeroValues(t *testing.T) {
	cfg := Default()
	zeroInt := 0
	zeroFloat := 0.0
	mode := ModeVercel
	merge(&cfg, Partial{
		Traffic: &PartialTraffic{MinPerMin: &zeroInt},
		Requests: &PartialRequest{
			DeviceRatio:  &zeroInt,
			UnknownRatio: &zeroInt,
			UniqueIPProb: &zeroFloat,
		},
		Schedule: &PartialSchedule{IdleOdds: &zeroFloat},
		Origin:   &PartialOrigin{Mode: &mode},
	})
	if cfg.Traffic.MinPerMin != 0 {
		t.Fatalf("MinPerMin = %d, want 0", cfg.Traffic.MinPerMin)
	}
	if cfg.Requests.DeviceRatio != 0 || cfg.Requests.UnknownRatio != 0 || cfg.Requests.UniqueIPProb != 0 {
		t.Fatalf("request zero values were not preserved: %+v", cfg.Requests)
	}
	if cfg.Schedule.IdleOdds != 0 {
		t.Fatalf("IdleOdds = %v, want 0", cfg.Schedule.IdleOdds)
	}
	if cfg.Origin.Mode != ModeVercel {
		t.Fatalf("Mode = %s, want vercel", cfg.Origin.Mode)
	}
}

func TestLoadPrecedence(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	t.Setenv("HOME", home)
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	globalDir := filepath.Join(home, ".hitmaker")
	if err := os.MkdirAll(globalDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(`{"traffic":{"minPerMin":2},"requests":{"deviceRatio":10}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(LocalConfigFile, []byte(`{"traffic":{"minPerMin":3},"requests":{"deviceRatio":0}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MIN_PER_MIN", "0")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Traffic.MinPerMin != 0 {
		t.Fatalf("env should win and preserve 0, got %d", cfg.Traffic.MinPerMin)
	}
	if cfg.Requests.DeviceRatio != 0 {
		t.Fatalf("local should win and preserve 0, got %d", cfg.Requests.DeviceRatio)
	}
}

func TestValidateRejectsUnknownBotSpec(t *testing.T) {
	cfg := Default()
	cfg.Requests.Bots = []string{"not-a-real-bot"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for unknown bot spec")
	}
}

func TestValidateAcceptsKnownBotSpec(t *testing.T) {
	cfg := Default()
	cfg.Requests.Bots = []string{"ai_crawler", "GPTBot"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	filter, err := cfg.BotFilter()
	if err != nil {
		t.Fatalf("BotFilter: %v", err)
	}
	if filter.Empty() {
		t.Fatal("filter should be constrained, not empty")
	}
}

func TestValidateAcceptsAutoMode(t *testing.T) {
	cfg := Default()
	cfg.Origin.Mode = ModeAuto
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEntropyDefaultsToChaos(t *testing.T) {
	cfg := Default()
	if cfg.Entropy.Level != EntropyChaos {
		t.Fatalf("default entropy level = %q, want chaos", cfg.Entropy.Level)
	}
	if !cfg.Entropy.Params().Active() {
		t.Fatal("default entropy should be active")
	}
}

func TestEntropyOffIsInert(t *testing.T) {
	cfg := Default()
	cfg.Entropy.Level = EntropyOff
	if cfg.Entropy.Params().Active() {
		t.Fatal("entropy off should be inert")
	}
	spread, breakout, viral := cfg.Entropy.EffectiveHuman()
	if spread != 0 || breakout != 0 || viral != 0 {
		t.Fatalf("entropy off knobs should be zero, got %d/%d/%d", spread, breakout, viral)
	}
}

func TestEntropyNamedLevelIgnoresStoredKnobs(t *testing.T) {
	cfg := Default()
	cfg.Entropy.Level = EntropyCalm
	cfg.Entropy.DeviceSpread = 999 // stale stored value must be ignored for named levels
	spread, _, _ := cfg.Entropy.EffectiveHuman()
	if spread != 12 {
		t.Fatalf("calm preset spread = %d, want 12", spread)
	}
}

func TestEntropyCustomUsesStoredKnobs(t *testing.T) {
	cfg := Default()
	cfg.Entropy = EntropyConfig{Level: EntropyCustom, DeviceSpread: 30, Breakout: 100, ViralPercent: 7}
	spread, breakout, viral := cfg.Entropy.EffectiveHuman()
	if spread != 30 || breakout != 100 || viral != 7 {
		t.Fatalf("custom knobs = %d/%d/%d, want 30/100/7", spread, breakout, viral)
	}
	params := cfg.Entropy.Params()
	if params.Sigma <= 1.19 || params.Sigma > 1.2 {
		t.Fatalf("breakout 100 should map to sigma ~1.2, got %v", params.Sigma)
	}
	if params.ViralOdds < 0.069 || params.ViralOdds > 0.071 {
		t.Fatalf("viral 7%% should map to odds ~0.07, got %v", params.ViralOdds)
	}
}

func TestValidateRejectsBadEntropy(t *testing.T) {
	cfg := Default()
	cfg.Entropy.Level = "bananas"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for unknown entropy level")
	}
	cfg = Default()
	cfg.Entropy.ViralPercent = 200
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for out-of-range viralPercent")
	}
}

func TestEntropyEnvOverride(t *testing.T) {
	t.Setenv("ENTROPY_LEVEL", "mayhem")
	t.Setenv("ENTROPY_VIRAL_PERCENT", "9")
	p, err := envPartial()
	if err != nil {
		t.Fatal(err)
	}
	if p.Entropy == nil || p.Entropy.Level == nil || *p.Entropy.Level != EntropyMayhem {
		t.Fatalf("env level not parsed: %+v", p.Entropy)
	}
	if p.Entropy.ViralPercent == nil || *p.Entropy.ViralPercent != 9 {
		t.Fatalf("env viral not parsed: %+v", p.Entropy)
	}
}

func TestBotsMergePreservesList(t *testing.T) {
	cfg := Default()
	bots := []string{"crawler"}
	merge(&cfg, Partial{Requests: &PartialRequest{Bots: &bots}})
	if len(cfg.Requests.Bots) != 1 || cfg.Requests.Bots[0] != "crawler" {
		t.Fatalf("bots = %v, want [crawler]", cfg.Requests.Bots)
	}
}
