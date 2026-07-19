package simulator

import (
	"fmt"
	"testing"

	"github.com/zeb-link/hitmaker/v2/internal/config"
)

func testTargets(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("https://example.test/link-%d", i)
	}
	return out
}

func TestPersonaOffIsIdentity(t *testing.T) {
	cfg := config.Default()
	cfg.Entropy.Level = config.EntropyOff
	cfg.Requests.DeviceRatio = 60
	cfg.Schedule.IdleOdds = 0.5
	personas := buildPersonas(cfg, testTargets(5), 12345)
	for target, p := range personas {
		if p.deviceRatio != 60 || p.energy != 1 || p.idleOdds != 0.5 || p.desync {
			t.Fatalf("entropy off should be identity for %s, got %+v", target, p)
		}
	}
}

func TestPersonaDeterministicForFixedSeed(t *testing.T) {
	cfg := config.Default() // chaos on
	targets := testTargets(20)
	a := buildPersonas(cfg, targets, 999)
	b := buildPersonas(cfg, targets, 999)
	for _, target := range targets {
		if a[target] != b[target] {
			t.Fatalf("same seed should reproduce persona for %s: %+v vs %+v", target, a[target], b[target])
		}
	}
}

func TestPersonaVariesAcrossLinks(t *testing.T) {
	cfg := config.Default() // chaos on
	cfg.Requests.DeviceRatio = 60
	personas := buildPersonas(cfg, testTargets(40), 4242)

	distinctDevice := map[int]struct{}{}
	sawBusier, sawQuieter := false, false
	for _, p := range personas {
		if p.deviceRatio < 0 || p.deviceRatio > 100 {
			t.Fatalf("device ratio out of range: %d", p.deviceRatio)
		}
		distinctDevice[p.deviceRatio] = struct{}{}
		if p.energy > 1.05 {
			sawBusier = true
		}
		if p.energy < 0.95 {
			sawQuieter = true
		}
	}
	if len(distinctDevice) < 5 {
		t.Fatalf("expected varied device ratios across links, saw %d distinct", len(distinctDevice))
	}
	if !sawBusier || !sawQuieter {
		t.Fatalf("expected a spread of energies (busier=%v quieter=%v)", sawBusier, sawQuieter)
	}
}

func TestPersonaViralHugsTopAndRarelyIdles(t *testing.T) {
	cfg := config.Default()
	cfg.Schedule.IdleOdds = 0.8
	// Custom level with every link forced viral, so the effect is deterministic.
	cfg.Entropy = config.EntropyConfig{
		Level:        config.EntropyCustom,
		DeviceSpread: 25,
		Breakout:     50,
		ViralPercent: 100,
	}
	personas := buildPersonas(cfg, testTargets(10), 7)
	for target, p := range personas {
		if p.energy < viralEnergyFloor {
			t.Fatalf("viral link %s energy %.2f below floor %.2f", target, p.energy, viralEnergyFloor)
		}
		if p.idleOdds > 0.8*0.1+1e-9 {
			t.Fatalf("viral link %s should rarely idle, got idleOdds %.3f", target, p.idleOdds)
		}
	}
}

func TestScaleRate(t *testing.T) {
	if got := scaleRate(10, 1); got != 10 {
		t.Fatalf("energy 1 should not change rate, got %d", got)
	}
	if got := scaleRate(10, 3); got != 30 {
		t.Fatalf("energy 3 should triple rate, got %d", got)
	}
	if got := scaleRate(10, 0); got != 1 {
		t.Fatalf("scaled rate must stay >= 1, got %d", got)
	}
}
