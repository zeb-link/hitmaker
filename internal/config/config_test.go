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

func TestBotsMergePreservesList(t *testing.T) {
	cfg := Default()
	bots := []string{"crawler"}
	merge(&cfg, Partial{Requests: &PartialRequest{Bots: &bots}})
	if len(cfg.Requests.Bots) != 1 || cfg.Requests.Bots[0] != "crawler" {
		t.Fatalf("bots = %v, want [crawler]", cfg.Requests.Bots)
	}
}
