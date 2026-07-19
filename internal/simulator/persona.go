package simulator

import (
	"hash/fnv"
	"math"
	"math/rand"

	"github.com/zeb-link/hitmaker/v2/internal/config"
)

// persona is a per-link personality drawn once at startup. When entropy is off
// it is the identity (base config values, energy 1, no offset), so every link
// behaves the same. When entropy is on, links differ: their device mix skews,
// their traffic volume varies on a long tail, busy links idle less, and a few
// become breakout "viral" links.
type persona struct {
	deviceRatio int     // this link's desktop share (0..100)
	energy      float64 // rate multiplier (median ~1, long right tail)
	idleOdds    float64 // this link's chance to idle after an active phase
	desync      bool    // stagger the first active phase so links don't idle in unison
}

// viralEnergyFloor is the minimum energy a breakout ("viral") link is pinned to,
// so it hugs the top of the rate range regardless of its own draw.
const viralEnergyFloor = 2.5

// buildPersonas assigns one persona per target. baseSeed makes the assignment
// reproducible when a run seed is set, and varies run-to-run otherwise.
func buildPersonas(cfg config.Config, targets []string, baseSeed int64) map[string]persona {
	params := cfg.Entropy.Params()
	out := make(map[string]persona, len(targets))
	for _, target := range targets {
		out[target] = makePersona(cfg, params, target, baseSeed)
	}
	return out
}

func makePersona(cfg config.Config, params config.EntropyParams, target string, baseSeed int64) persona {
	base := persona{
		deviceRatio: cfg.Requests.DeviceRatio,
		energy:      1,
		idleOdds:    cfg.Schedule.IdleOdds,
	}
	if !params.Active() {
		return base
	}
	rng := rand.New(rand.NewSource(personaSeed(baseSeed, target)))

	// Audience mix: skew this link's desktop share around the base, so the
	// device breakdown differs per link instead of every link sitting at the
	// same ratio.
	if params.DeviceSpread > 0 {
		wander := int(math.Round(float64(params.DeviceSpread) * (2*rng.Float64() - 1)))
		base.deviceRatio = clampInt(cfg.Requests.DeviceRatio+wander, 0, 100)
	}

	// Volume: a log-normal multiplier with median 1 and a natural long tail —
	// most links modest, a few genuinely busy.
	if params.Sigma > 0 {
		base.energy = math.Exp(params.Sigma * rng.NormFloat64())
	}

	// Busier links take fewer idle breaks; quieter links nap more.
	if base.energy > 0 {
		base.idleOdds = clampFloat(cfg.Schedule.IdleOdds/base.energy, 0, 1)
	}

	// Kingmaker: a small share of links become breakouts — they hug the top of
	// the rate range and rarely idle.
	if params.ViralOdds > 0 && rng.Float64() < params.ViralOdds {
		if base.energy < viralEnergyFloor {
			base.energy = viralEnergyFloor
		}
		base.idleOdds = clampFloat(cfg.Schedule.IdleOdds*0.1, 0, 1)
	}

	// Traffic starts immediately, but the first active phase is cut short by a
	// random amount so links reach their idle roll at different times instead of
	// all going quiet at once.
	base.desync = true
	return base
}

func personaSeed(baseSeed int64, target string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(target))
	return baseSeed ^ int64(h.Sum64())
}

// scaleRate applies a persona's energy multiplier to a base per-minute rate,
// keeping at least 1 hit/min.
func scaleRate(rate int, energy float64) int {
	scaled := int(math.Round(float64(rate) * energy))
	if scaled < 1 {
		scaled = 1
	}
	return scaled
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func clampFloat(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
