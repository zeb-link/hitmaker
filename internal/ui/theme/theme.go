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

	// Semantic, kept distinct from the accent: success vs failure.
	Emerald = lipgloss.Color("#4cc38a") // hits / success
	Tomato  = lipgloss.Color("#ec6142") // errors / miss

	// Retro banner ramp — a warm rainbow nod for the intro wordmark only.
	HotPink = lipgloss.Color("#e0857a") // coral
	Cyan    = lipgloss.Color("#6fb3a8") // dim teal — the one cool note
	Mint    = lipgloss.Color("#9fc08a") // sage
	Amber   = Accent
	Red     = Tomato
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

	// Tick is the left accent marker used to indicate the focused row. It is a
	// single amber cell — quiet, and (unlike a width-padded background) it can
	// never wrap the row.
	Tick = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	// SelectedRow fills a row background softly. Retained for any surface that
	// still wants a filled selection; the deck and table use Tick instead.
	SelectedRow = lipgloss.NewStyle().Foreground(Ink).Background(Panel)
)
