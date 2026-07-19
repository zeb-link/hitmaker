// Package theme defines Hitmaker's terminal visual vocabulary.
package theme

import "charm.land/lipgloss/v2"

// lipgloss v2 colors are color.Color values (not string constants), so these
// are var, not const.
var (
	HotPink   = lipgloss.Color("205")
	Laser     = lipgloss.Color("198")
	Mint      = lipgloss.Color("121")
	Cyan      = lipgloss.Color("51")
	Amber     = lipgloss.Color("214")
	Red       = lipgloss.Color("203")
	Ink       = lipgloss.Color("230")
	Muted     = lipgloss.Color("244")
	Dim       = lipgloss.Color("238")
	Panel     = lipgloss.Color("235")
	Highlight = lipgloss.Color("54") // selected-row background (deep violet)
)

var (
	Logo        = lipgloss.NewStyle().Bold(true).Foreground(HotPink)
	Title       = lipgloss.NewStyle().Bold(true).Foreground(Ink)
	Subtle      = lipgloss.NewStyle().Foreground(Muted)
	Command     = lipgloss.NewStyle().Foreground(Cyan)
	Good        = lipgloss.NewStyle().Foreground(Mint)
	Warn        = lipgloss.NewStyle().Foreground(Amber)
	Bad         = lipgloss.NewStyle().Foreground(Red)
	Focus       = lipgloss.NewStyle().Foreground(HotPink).Bold(true)
	Border      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Dim).Padding(0, 1)
	FocusBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(HotPink).Padding(0, 1)
	Pill        = lipgloss.NewStyle().Foreground(Ink).Background(Dim).Padding(0, 1)
	PillHot     = lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(HotPink).Bold(true).Padding(0, 1)

	// SelectedRow highlights the whole focused row so navigation is obvious.
	SelectedRow = lipgloss.NewStyle().Foreground(Ink).Background(Highlight).Bold(true)
)
