// Package theme defines Hitmaker's terminal visual vocabulary.
//
// "Ember" direction: a warm ground, a single amber accent, and soft
// emerald/tomato semantics for hits vs errors. Focus is quiet — a left accent
// tick, never a full neon bar. A warm rainbow ramp survives only in the intro
// banner, as a deliberate retro nod.
//
// Both a dark and a light ground ship. Every role is a light/dark pair, and
// Configure picks the right side once the terminal background is known. Until
// then we default to the dark ground (see init). Detection is wired at the
// edges: the Bubble Tea models listen for tea.BackgroundColorMsg; the plain
// CLI prints use lipgloss.HasDarkBackground. Nothing here reads the terminal —
// callers hand us the answer.
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Colors resolved for the active ground. Configure assigns these.
var (
	Ink   color.Color // primary text
	Muted color.Color // secondary text
	Dim   color.Color // borders, inactive, empty track
	Panel color.Color // raised surface / focused-param ground

	Accent    color.Color // the one bold color — focus/active is amber
	AccentDim color.Color // soft amber wash for active chips
	Gold      color.Color // brighter amber, for the intro shimmer
	Shadow    color.Color // drop-shadow ground for floating panels

	Emerald color.Color // hits / success
	Tomato  color.Color // errors / miss

	Amber color.Color // aliases Accent for the intro shimmer's base tone
)

// Styles resolved for the active ground. Configure rebuilds these from the
// colors above.
var (
	Logo    lipgloss.Style
	Title   lipgloss.Style
	Subtle  lipgloss.Style
	Command lipgloss.Style
	Good    lipgloss.Style
	Warn    lipgloss.Style
	Bad     lipgloss.Style
	Focus   lipgloss.Style

	Border      lipgloss.Style
	FocusBorder lipgloss.Style

	// Chips: inactive reads as a quiet outline (muted text on the panel ground);
	// active gains amber over a soft amber wash — no harsh saturated block.
	Pill    lipgloss.Style
	PillHot lipgloss.Style

	// Live is the dashboard's running indicator — bright emerald on a dark
	// emerald ground (and the mirror on a light ground).
	Live lipgloss.Style

	// Tick is the left accent marker used to indicate the focused row. It is a
	// single amber cell — quiet, and (unlike a width-padded background) it can
	// never wrap the row.
	Tick lipgloss.Style

	// SelectedRow fills a row background softly. Retained for any surface that
	// still wants a filled selection; the deck and table use Tick instead.
	SelectedRow lipgloss.Style

	// keyCap / keyLabel / keyChip back the shared KeyHint chip: a bright accent
	// key and a muted label on a muted ground. Defined once so the dashboard
	// footer, the config command bar, and the save-dialog buttons all match.
	// Each segment carries the Panel ground so the chip fill stays continuous
	// across the key→label color change (an inner reset would otherwise drop it).
	keyCap   lipgloss.Style
	keyLabel lipgloss.Style
	keyChip  lipgloss.Style
)

// dark is the default until the terminal background is detected, so a terminal
// we never get to query still renders the original Ember ground.
func init() { Configure(true) }

// Configure resolves every color and style for the given ground. Pass true for
// a dark terminal, false for a light one. Safe to call again when the detected
// background changes (Bubble Tea can re-emit it) — every View reads these vars
// fresh, so the next frame adapts.
func Configure(isDark bool) {
	// pick(light, dark) — light color first, per lipgloss.LightDarkFunc.
	pick := lipgloss.LightDark(isDark)

	Ink = pick(lipgloss.Color("#2c2620"), lipgloss.Color("#e7e0d6"))
	Muted = pick(lipgloss.Color("#6f6456"), lipgloss.Color("#8a7f72"))
	Dim = pick(lipgloss.Color("#cbbfa8"), lipgloss.Color("#4a423b"))
	// Panel is the chip/footer fill. On a white terminal it must sit clearly
	// below the page, or the shortcut chips read as loose floating text — so the
	// light ground is a warm sand, not a near-white wash.
	Panel = pick(lipgloss.Color("#e7dbc4"), lipgloss.Color("#241d16"))

	Accent = pick(lipgloss.Color("#9c6510"), lipgloss.Color("#d9a441"))
	// AccentDim is the active-chip wash. A touch more amber than the panel so a
	// selected pill reads as lit, not just filled.
	AccentDim = pick(lipgloss.Color("#f2dca8"), lipgloss.Color("#3a2c17"))
	Gold = pick(lipgloss.Color("#c8901d"), lipgloss.Color("#f0d68c"))
	Shadow = pick(lipgloss.Color("#d6ccbb"), lipgloss.Color("#07060a"))

	Emerald = pick(lipgloss.Color("#1c8f5a"), lipgloss.Color("#4cc38a"))
	Tomato = pick(lipgloss.Color("#c53a1f"), lipgloss.Color("#ec6142"))

	Amber = Accent

	// Live's own bright-on-soft emerald pair, mirrored for the light ground.
	liveFg := pick(lipgloss.Color("#10704a"), lipgloss.Color("#5fd39b"))
	liveBg := pick(lipgloss.Color("#d5efe1"), lipgloss.Color("#123528"))

	Logo = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	Title = lipgloss.NewStyle().Bold(true).Foreground(Ink)
	Subtle = lipgloss.NewStyle().Foreground(Muted)
	Command = lipgloss.NewStyle().Foreground(Accent)
	Good = lipgloss.NewStyle().Foreground(Emerald)
	Warn = lipgloss.NewStyle().Foreground(Accent)
	Bad = lipgloss.NewStyle().Foreground(Tomato)
	Focus = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	Border = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Dim).Padding(0, 1)
	FocusBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Accent).Padding(0, 1)

	Pill = lipgloss.NewStyle().Foreground(Muted).Background(Panel).Padding(0, 1)
	PillHot = lipgloss.NewStyle().Foreground(Accent).Background(AccentDim).Bold(true).Padding(0, 1)

	Live = lipgloss.NewStyle().Foreground(liveFg).Background(liveBg).Bold(true).Padding(0, 1)

	Tick = lipgloss.NewStyle().Foreground(Accent).Bold(true)
	SelectedRow = lipgloss.NewStyle().Foreground(Ink).Background(Panel)

	keyCap = lipgloss.NewStyle().Foreground(Accent).Background(Panel).Bold(true)
	keyLabel = lipgloss.NewStyle().Foreground(Muted).Background(Panel)
	keyChip = lipgloss.NewStyle().Background(Panel).Padding(0, 1)
}

// KeyHint renders one keyboard-shortcut chip — a brighter accent key followed by
// a muted label, on the muted chip ground. This is the single source of truth for
// every shortcut/hint bar in the app.
func KeyHint(key, label string) string {
	body := keyCap.Render(key)
	if label != "" {
		body += keyLabel.Render(" " + label)
	}
	return keyChip.Render(body)
}
