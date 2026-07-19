package tui

import (
	"time"

	"charm.land/bubbles/v2/spinner"

	"github.com/zeb-link/hitmaker/v2/internal/config"
	"github.com/zeb-link/hitmaker/v2/internal/simulator"
)

// SampleDashboard renders one dashboard frame from representative fixture data,
// with no runner or network. It exists so documentation screenshots can be
// regenerated deterministically (see the hidden `frame` command).
func SampleDashboard(width, height int) string {
	m := Model{width: width, height: height, selected: 0, snapshot: sampleSnapshot()}
	m.spinner = spinner.New()
	// A plain glyph renders cleanly in screenshot fonts (the live braille
	// spinner animates fine in a real terminal but shows as tofu in images).
	m.spinner.Spinner = spinner.Spinner{Frames: []string{"●"}}
	return m.dashboardView()
}

// SampleConfig renders one config-editor frame from the default config.
func SampleConfig(width, height int) string {
	e := newConfigEditor(config.Default())
	e.focus = 5 // Bot pool
	_, rightWidth := dashboardPaneWidths(width)
	e = e.WithHelpWidth(rightWidth)
	return e.View(width, height, nil)
}

func sampleSnapshot() simulator.Snapshot {
	base := time.Date(2026, 7, 7, 3, 0, 0, 0, time.UTC)
	s := simulator.Snapshot{
		Uptime:      104 * time.Second,
		TotalHits:   248,
		TotalErrors: 2,
		WorkerCount: 3,
		Targets: []simulator.TargetStats{
			{Target: "https://ze.bra/launch", Hits: 141, CurrentRate: 22, Errors: 1, ActiveWorkers: 1},
			{Target: "https://ze.bra/promo-qr", Hits: 84, CurrentRate: 15, ActiveWorkers: 1},
			{Target: "https://ze.bra/newsletter", Hits: 23, CurrentRate: 0, Errors: 1, ActiveWorkers: 0},
		},
	}
	statuses := []int{200, 200, 302, 200, 200, 302, 200, 429, 200, 200, 302, 200, 200, 200, 302, 200, 200, 302, 200, 200}
	names := []string{"ze.bra/launch", "ze.bra/promo-qr", "ze.bra/launch", "ze.bra/newsletter"}
	for i := 0; i < len(statuses); i++ {
		st := statuses[i]
		hit := simulator.HitResult{
			Target:   "https://" + names[i%len(names)],
			WorkerID: (i % 3) + 1,
			Status:   st,
			At:       base.Add(time.Duration(-i*4) * time.Second),
		}
		if st >= 500 {
			hit.Err = "server error"
		}
		s.Recent = append(s.Recent, hit)
	}
	return s
}
