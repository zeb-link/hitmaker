package tui

import (
	"strings"
	"testing"
	"time"

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
		// Well past the reveal so the wordmark has fully burned in to amber.
		introStart: time.Now().Add(-3 * time.Second),
		introUntil: time.Now().Add(-2 * time.Second),
	}
	// lipgloss v2 emits color escapes even off a TTY, so strip ANSI before
	// asserting on the visible glyphs.
	plain := stripANSI(m.introView())
	if !strings.ContainsAny(plain, "█▀▄") {
		t.Fatalf("intro missing block wordmark:\n%s", plain)
	}
	if !strings.Contains(plain, "MAKING THE HITS") {
		t.Fatalf("intro missing subtitle:\n%s", plain)
	}
}

func TestIntroBannerRendersAtEveryFrame(t *testing.T) {
	for _, ms := range []int{0, 40, 150, 320, 600, 1200} {
		m := Model{
			width:      100,
			height:     32,
			introStart: time.Now().Add(-time.Duration(ms) * time.Millisecond),
			introUntil: time.Now().Add(time.Second),
		}
		out := m.animatedBanner()
		if lines := strings.Count(out, "\n") + 1; lines != len(hitmakerBanner) {
			t.Fatalf("banner at %dms rendered %d rows, want %d", ms, lines, len(hitmakerBanner))
		}
		// The whole intro (starfield + wordmark + subtitle) must always compose.
		if strings.TrimSpace(stripANSI(m.introView())) == "" {
			t.Fatalf("intro produced empty output at %dms", ms)
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
