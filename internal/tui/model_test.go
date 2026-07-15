package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/zeb-link/hitmaker/v2/internal/config"
)

func TestDashboardPaneWidthsMatchesConfigHelpWidth(t *testing.T) {
	width := 160
	contentWidth := width - bodyInsetX*2
	left, right := dashboardPaneWidths(contentWidth)
	if right < 34 || right > 60 {
		t.Fatalf("right width = %d, want clamped recent pane width", right)
	}
	editor := newConfigEditor(config.Default()).WithHelpWidth(right)
	gotLeft, gotRight := editor.editorColumnWidths(contentWidth)
	if gotRight != right {
		t.Fatalf("config help width = %d, want dashboard right width %d", gotRight, right)
	}
	if gotLeft+gotRight+1 != contentWidth {
		t.Fatalf("config columns %d + %d do not fill content width %d", gotLeft, gotRight, contentWidth)
	}
	if left+right+1 != contentWidth {
		t.Fatalf("dashboard columns %d + %d do not fill content width %d", left, right, contentWidth)
	}
}

func TestIntroViewRendersAnimatedBanner(t *testing.T) {
	m := Model{
		width:      100,
		height:     32,
		introStart: time.Now().Add(-120 * time.Millisecond),
		introUntil: time.Now().Add(780 * time.Millisecond),
	}
	m.spinner = spinner.New()
	m.spinner.Spinner = spinner.Spinner{Frames: []string{"●"}}
	view := m.introView()
	if !strings.Contains(view, "█████") {
		t.Fatalf("intro missing block banner:\n%s", view)
	}
	if !strings.Contains(view, "synthetic traffic engine") {
		t.Fatalf("intro missing subtitle:\n%s", view)
	}
	if !strings.Contains(view, "warming up") {
		t.Fatalf("intro missing loading text:\n%s", view)
	}
}
