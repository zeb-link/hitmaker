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

// SampleIntro renders one frame of the intro — caught mid burn-in, so the whole
// wordmark is up and still glowing hot at the leading edge — for a splash shot.
func SampleIntro(width, height int) string {
	m := Model{width: width, height: height, introStart: time.Now().Add(-300 * time.Millisecond)}
	return m.introView()
}

func sampleSnapshot() simulator.Snapshot {
	base := time.Date(2026, 7, 7, 3, 0, 0, 0, time.UTC)
	s := simulator.Snapshot{
		Uptime:      6*time.Minute + 12*time.Second,
		TotalHits:   1198,
		TotalErrors: 3,
		WorkerCount: 7,
		Targets: []simulator.TargetStats{
			{Target: "https://z.dk/launch", Hits: 412, CurrentRate: 34, ActiveWorkers: 2},
			{Target: "https://z.dk/promo-qr", Hits: 288, CurrentRate: 22, ActiveWorkers: 1},
			{Target: "https://z.dk/newsletter", Hits: 176, CurrentRate: 12, Errors: 1, ActiveWorkers: 1},
			{Target: "https://z.dk/podcast", Hits: 141, CurrentRate: 18, ActiveWorkers: 1},
			{Target: "https://z.dk/blackfriday", Hits: 97, CurrentRate: 9, ActiveWorkers: 1},
			{Target: "https://z.dk/docs", Hits: 63, CurrentRate: 0, ActiveWorkers: 0},
			{Target: "https://z.dk/careers", Hits: 21, CurrentRate: 4, Errors: 2, ActiveWorkers: 1},
		},
	}
	statuses := []int{200, 200, 302, 200, 200, 302, 200, 429, 200, 200, 302, 200, 200, 200, 302, 200}
	names := []string{
		"z.dk/launch", "z.dk/promo-qr", "z.dk/podcast", "z.dk/launch",
		"z.dk/blackfriday", "z.dk/newsletter", "z.dk/careers", "z.dk/launch",
	}
	for i := 0; i < len(statuses); i++ {
		st := statuses[i]
		hit := simulator.HitResult{
			Target:   "https://" + names[i%len(names)],
			WorkerID: (i % 7) + 1,
			Status:   st,
			At:       base.Add(time.Duration(-i*3) * time.Second),
		}
		if st >= 500 {
			hit.Err = "server error"
		}
		s.Recent = append(s.Recent, hit)
	}
	return s
}
