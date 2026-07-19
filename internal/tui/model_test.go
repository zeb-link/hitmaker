package tui

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/spinner"
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
		width:  100,
		height: 32,
		// Well past the animation so every variant has resolved to the solid
		// wordmark, whichever one was picked.
		introStart: time.Now().Add(-3 * time.Second),
		introUntil: time.Now().Add(-2 * time.Second),
	}
	m.spinner = spinner.New()
	m.spinner.Spinner = spinner.Spinner{Frames: []string{"●"}}
	// lipgloss v2 emits color escapes even off a TTY, and the banner is rendered
	// one glyph-cell at a time for the rainbow animation — so the block run is
	// split by escape codes. Strip ANSI before asserting the visible shape.
	plain := stripANSI(m.introView())
	if !strings.Contains(plain, "█████") {
		t.Fatalf("intro missing block banner:\n%s", plain)
	}
	if !strings.Contains(plain, "synthetic traffic engine") {
		t.Fatalf("intro missing subtitle:\n%s", plain)
	}
	if !strings.Contains(plain, "warming up") {
		t.Fatalf("intro missing loading text:\n%s", plain)
	}
}

func TestIntroVariantsRenderAtEveryFrame(t *testing.T) {
	for v := 0; v < introVariantCount; v++ {
		for _, ms := range []int{0, 40, 120, 300, 600, 1200} {
			m := Model{
				width:        100,
				height:       32,
				introVariant: v,
				introStart:   time.Now().Add(-time.Duration(ms) * time.Millisecond),
				introUntil:   time.Now().Add(time.Second),
			}
			m.spinner = spinner.New()
			m.spinner.Spinner = spinner.Spinner{Frames: []string{"●"}}
			out := m.animatedBanner()
			if strings.TrimSpace(stripANSI(out)) == "" {
				t.Fatalf("variant %d produced an empty banner at %dms", v, ms)
			}
			if lines := strings.Count(out, "\n") + 1; lines != len(hitmakerBanner) {
				t.Fatalf("variant %d at %dms rendered %d rows, want %d", v, ms, lines, len(hitmakerBanner))
			}
		}
	}
}

// stripANSI removes SGR/escape sequences so tests can assert on visible glyphs.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inEsc = false
			}
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}
