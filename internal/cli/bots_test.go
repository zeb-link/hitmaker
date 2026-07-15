package cli

import (
	"testing"

	"github.com/zeb-link/hitmaker/internal/config"
)

func TestApplyBotFlagsImpliesFullBotTraffic(t *testing.T) {
	cfg := config.Default()
	cfg.Requests.UnknownRatio = 5
	if err := applyBotFlags(&cfg, "ai", 0, false); err != nil {
		t.Fatalf("applyBotFlags: %v", err)
	}
	if cfg.Requests.UnknownRatio != 100 {
		t.Fatalf("unknownRatio = %d, want 100 (implied by --bots)", cfg.Requests.UnknownRatio)
	}
	if len(cfg.Requests.Bots) != 1 || cfg.Requests.Bots[0] != "ai" {
		t.Fatalf("bots = %v, want [ai]", cfg.Requests.Bots)
	}
}

func TestApplyBotFlagsRespectsExplicitRatio(t *testing.T) {
	cfg := config.Default()
	if err := applyBotFlags(&cfg, "crawler", 40, true); err != nil {
		t.Fatalf("applyBotFlags: %v", err)
	}
	if cfg.Requests.UnknownRatio != 40 {
		t.Fatalf("unknownRatio = %d, want 40", cfg.Requests.UnknownRatio)
	}
}

func TestApplyBotFlagsRejectsUnknownBot(t *testing.T) {
	cfg := config.Default()
	if err := applyBotFlags(&cfg, "nope", 0, false); err == nil {
		t.Fatal("expected error for unknown bot token")
	}
}

func TestApplyBotFlagsRejectsBadRatio(t *testing.T) {
	cfg := config.Default()
	if err := applyBotFlags(&cfg, "", 150, true); err == nil {
		t.Fatal("expected error for out-of-range ratio")
	}
}
