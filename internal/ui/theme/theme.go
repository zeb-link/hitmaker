// Package theme defines Hitmaker's terminal visual vocabulary.
//
// "Ember" direction: a warm near-black ground, a single amber accent, and soft
// emerald/tomato semantics for hits vs errors. Focus is quiet — a left accent
// tick, never a full neon bar. A warm rainbow ramp survives only in the intro
// banner, as a deliberate retro nod.
package theme

import "charm.land/lipgloss/v2"

// lipgloss v2 colors are color.Color values (not string constants), so these
// are var, not const.
var (
	// Core neutrals, warm-biased so the amber accent belongs to the palette.
	Ink   = lipgloss.Color("#e7e0d6") // primary text
	Muted = lipgloss.Color("#8a7f72") // secondary text
	Dim   = lipgloss.Color("#4a423b") // borders, inactive, empty track
	Panel = lipgloss.Color("#241d16") // raised surface / focused-param ground

	// Accent — the one bold color. Everything focus/active is amber.
	Accent    = lipgloss.Color("#d9a441")
	AccentDim = lipgloss.Color("#3a2c17") // soft amber wash for active chips
	Gold      = lipgloss.Color("#f0d68c") // brighter amber, for the intro shimmer
	Shadow    = lipgloss.Color("#07060a") // drop-shadow ground for floating panels

	// Semantic, kept distinct from the accent: success vs failure.
	Emerald = lipgloss.Color("#4cc38a") // hits / success
	Tomato  = lipgloss.Color("#ec6142") // errors / miss

	// Amber aliases the accent for the intro shimmer's base tone.
	Amber = Accent
)

var (
	Logo    = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	Title   = lipgloss.NewStyle().Bold(true).Foreground(Ink)
	Subtle  = lipgloss.NewStyle().Foreground(Muted)
	Command = lipgloss.NewStyle().Foreground(Accent)
	Good    = lipgloss.NewStyle().Foreground(Emerald)
	Warn    = lipgloss.NewStyle().Foreground(Accent)
	Bad     = lipgloss.NewStyle().Foreground(Tomato)
	Focus   = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	Border      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Dim).Padding(0, 1)
	FocusBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Accent).Padding(0, 1)

	// Chips: inactive reads as a quiet outline (muted text on the panel ground);
	// active gains amber over a soft amber wash — no harsh saturated block.
	Pill    = lipgloss.NewStyle().Foreground(Muted).Background(Panel).Padding(0, 1)
	PillHot = lipgloss.NewStyle().Foreground(Accent).Background(AccentDim).Bold(true).Padding(0, 1)

	// Live is the dashboard's running indicator — bright emerald on a dark
	// emerald ground.
	Live = lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd39b")).Background(lipgloss.Color("#123528")).Bold(true).Padding(0, 1)

	// keyCap / keyLabel / keyChip back the shared KeyHint chip: a bright accent
	// key and a muted label on a muted ground. Defined once so the dashboard
	// footer, the config command bar, and the save-dialog buttons all match.
	// Each segment carries the Panel ground so the chip fill stays continuous
	// across the key→label color change (an inner reset would otherwise drop it).
	keyCap   = lipgloss.NewStyle().Foreground(Accent).Background(Panel).Bold(true)
	keyLabel = lipgloss.NewStyle().Foreground(Muted).Background(Panel)
	keyChip  = lipgloss.NewStyle().Background(Panel).Padding(0, 1)
)

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

var (
	// Tick is the left accent marker used to indicate the focused row. It is a
	// single amber cell — quiet, and (unlike a width-padded background) it can
	// never wrap the row.
	Tick = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	// SelectedRow fills a row background softly. Retained for any surface that
	// still wants a filled selection; the deck and table use Tick instead.
	SelectedRow = lipgloss.NewStyle().Foreground(Ink).Background(Panel)
)
