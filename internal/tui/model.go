package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/zeb-link/hitmaker/v2/internal/config"
	"github.com/zeb-link/hitmaker/v2/internal/simulator"
	"github.com/zeb-link/hitmaker/v2/internal/ui/theme"
)

type Options struct {
	Config  config.Config
	Targets []string
	NoIntro bool
}

type mode int

const (
	modeDashboard mode = iota
	modeConfig
)

type tickMsg time.Time

const (
	bodyInsetX = 2
	bodyInsetY = 1
)

type Model struct {
	cfg     config.Config
	targets []string
	runner  *simulator.Runner

	mode       mode
	width      int
	height     int
	selected   int
	scroll     int
	snapshot   simulator.Snapshot
	spinner    spinner.Model
	configEdit configEditor
	introStart time.Time
	introUntil time.Time
	err        error
}

func New(opts Options) (Model, error) {
	runner, err := simulator.New(context.Background(), simulator.Options{Config: opts.Config, Targets: opts.Targets})
	if err != nil {
		return Model{}, err
	}
	spin := spinner.New()
	spin.Spinner = spinner.Spinner{Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}, FPS: 80 * time.Millisecond}
	m := Model{
		cfg:        opts.Config,
		targets:    opts.Targets,
		runner:     runner,
		spinner:    spin,
		configEdit: newConfigEditor(opts.Config),
	}
	if !opts.NoIntro {
		m.introStart = time.Now()
		m.introUntil = m.introStart.Add(900 * time.Millisecond)
	}
	return m, nil
}

func (m Model) Init() tea.Cmd {
	m.runner.Start()
	return tea.Batch(m.spinner.Tick, tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case tickMsg:
		m.snapshot = m.runner.Snapshot()
		cmds = append(cmds, tick())
	case tea.KeyPressMsg:
		if m.mode == modeConfig {
			next, action, cmd := m.configEdit.Update(msg)
			m.configEdit = next
			cmds = append(cmds, cmd)
			switch action {
			case configActionClose:
				m.mode = modeDashboard
			case configActionApply:
				// Save & close: persist locally by default, hot-reload the live
				// runner with the new config, and drop back to the dashboard so
				// the effect is immediately visible.
				m.cfg = m.configEdit.cfg
				m.runner.StopAndWait(2 * time.Second)
				runner, err := simulator.New(context.Background(), simulator.Options{Config: m.cfg, Targets: m.targets})
				if err != nil {
					m.err = err
				} else {
					m.runner = runner
					m.runner.Start()
					m.err = config.SaveLocal(m.cfg)
					m.mode = modeDashboard
				}
			case configActionSaveGlobal:
				m.err = config.SaveGlobal(m.configEdit.cfg)
			case configActionSaveLocal:
				m.err = config.SaveLocal(m.configEdit.cfg)
			case configActionDefaults:
				m.configEdit = newConfigEditor(config.Default())
			}
			return m, tea.Batch(cmds...)
		}
		switch keyString(msg) {
		case "ctrl+c", "q":
			m.runner.StopAndWait(2 * time.Second)
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.snapshot.Targets)-1 {
				m.selected++
			}
		case "K":
			if len(m.snapshot.Targets) > 0 {
				m.runner.TogglePause(m.snapshot.Targets[m.selected].Target)
			}
		case "c", "C":
			m.configEdit = newConfigEditor(m.cfg)
			m.mode = modeConfig
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("")
	}
	var content string
	switch {
	case m.mode == modeConfig:
		contentWidth := max(1, m.width-bodyInsetX*2)
		_, rightWidth := dashboardPaneWidths(contentWidth)
		content = m.configEdit.WithHelpWidth(rightWidth).View(m.width, m.height, m.err)
	case time.Now().Before(m.introUntil):
		content = m.introView()
	default:
		content = m.dashboardView()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// keyString maps a v2 key press to the string the dashboard switches on:
// printable keys use their literal text (so "K" and "k" stay distinct), and
// named keys (up, enter, ctrl+c) fall back to the canonical String() form.
func keyString(msg tea.KeyPressMsg) string {
	if t := msg.Key().Text; t != "" {
		return t
	}
	return msg.String()
}

// bannerGlyphs are 5-row-tall, 5-cell-wide block letters. Composing the wordmark
// from these guarantees the columns stay aligned no matter which letters change.
var bannerGlyphs = map[rune][5]string{
	'H': {"█   █", "█   █", "█████", "█   █", "█   █"},
	'I': {"█████", "  █  ", "  █  ", "  █  ", "█████"},
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'M': {"█   █", "██ ██", "█ █ █", "█   █", "█   █"},
	'A': {" ███ ", "█   █", "█████", "█   █", "█   █"},
	'K': {"█   █", "█  █ ", "███  ", "█  █ ", "█   █"},
	'E': {"█████", "█    ", "████ ", "█    ", "█████"},
	'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
}

func buildBanner(word string) []string {
	rows := make([]string, 5)
	for i := range rows {
		parts := make([]string, 0, len(word))
		for _, r := range word {
			g := bannerGlyphs[r]
			parts = append(parts, g[i])
		}
		rows[i] = strings.Join(parts, " ")
	}
	return rows
}

var hitmakerBanner = buildBanner("HITMAKER")

func (m Model) introView() string {
	var head string
	if m.width >= 50 {
		head = m.animatedBanner()
	} else {
		head = m.animatedIntroText("H I T M A K E R")
	}
	body := lipgloss.JoinVertical(lipgloss.Center,
		head,
		"",
		theme.Subtle.Render("synthetic traffic engine"),
		"",
		theme.Focus.Render(m.spinner.View()+" warming up"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

// animatedBanner paints the block wordmark in amber with a gold shimmer sweeping
// left to right, led by an emerald crest — cohesive with the Ember palette
// (amber + gold + a single complementary green) rather than a rainbow.
func (m Model) animatedBanner() string {
	elapsed := time.Since(m.introStart)
	if m.introStart.IsZero() {
		elapsed = 0
	}
	// The wave sweeps across the wordmark once over the intro, with a short
	// lead-in from off the left edge.
	wave := int(elapsed/(20*time.Millisecond)) - 4
	base := lipgloss.NewStyle().Foreground(theme.Amber).Bold(true)
	glow := lipgloss.NewStyle().Foreground(theme.Gold).Bold(true)
	crest := lipgloss.NewStyle().Foreground(theme.Emerald).Bold(true)
	rows := make([]string, len(hitmakerBanner))
	for row, line := range hitmakerBanner {
		var b strings.Builder
		for col, r := range line {
			if r == ' ' {
				b.WriteByte(' ')
				continue
			}
			switch d := col - wave; {
			case d == 0:
				b.WriteString(crest.Render(string(r)))
			case d >= -2 && d <= 1:
				b.WriteString(glow.Render(string(r)))
			default:
				b.WriteString(base.Render(string(r)))
			}
		}
		rows[row] = b.String()
	}
	return strings.Join(rows, "\n")
}

func (m Model) animatedIntroText(text string) string {
	elapsed := time.Since(m.introStart)
	if m.introStart.IsZero() {
		elapsed = 0
	}
	wave := int(elapsed/(20*time.Millisecond)) - 2
	base := lipgloss.NewStyle().Foreground(theme.Amber).Bold(true)
	glow := lipgloss.NewStyle().Foreground(theme.Gold).Bold(true)
	crest := lipgloss.NewStyle().Foreground(theme.Emerald).Bold(true)
	parts := make([]string, 0, len(text))
	for i, r := range text {
		if r == ' ' {
			parts = append(parts, " ")
			continue
		}
		switch d := i - wave; {
		case d == 0:
			parts = append(parts, crest.Render(string(r)))
		case d >= -1 && d <= 1:
			parts = append(parts, glow.Render(string(r)))
		default:
			parts = append(parts, base.Render(string(r)))
		}
	}
	return strings.Join(parts, "")
}

func (m Model) dashboardView() string {
	snap := m.snapshot
	header := m.headerView(snap)
	footer := m.footerView()
	// Body gets whatever is left after the pinned header and footer. Using the
	// actual rendered heights means a wrapping header can't push the footer off
	// screen — the body simply shrinks. Everything below is clamped to bodyHeight
	// so the recent-hits list can never overflow the viewport.
	bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	contentWidth := max(1, m.width-bodyInsetX*2)
	contentHeight := max(1, bodyHeight-bodyInsetY)
	leftWidth, rightWidth := dashboardPaneWidths(contentWidth)

	table := m.tableView(snap, leftWidth, contentHeight)
	var main string
	if rightWidth >= 34 {
		recent := m.recentView(snap, rightWidth, contentHeight)
		main = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(leftWidth).Render(table),
			" ",
			recent,
		)
	} else {
		main = table
	}
	main = insetBlock(main, bodyInsetX, bodyInsetY)
	main = lipgloss.NewStyle().Width(m.width).Height(bodyHeight).MaxHeight(bodyHeight).Render(main)
	return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
}

func dashboardPaneWidths(width int) (int, int) {
	if width < 92 {
		return width, 0
	}
	// Cap the recent pane so it stays a companion column instead of eating half
	// of an ultrawide terminal; the main table/config deck takes the rest.
	rightWidth := clampInt(int(float64(width)*0.40), 34, 60)
	leftWidth := width - rightWidth - 1
	return leftWidth, rightWidth
}

func (m Model) headerView(snap simulator.Snapshot) string {
	status := theme.PillHot.Render(m.spinner.View() + " LIVE")
	if snap.WorkerCapHit {
		status += " " + theme.Warn.Render("cap")
	}
	hits := theme.Good.Render(fmt.Sprintf("%d", snap.TotalHits))
	errs := theme.Bad.Render(fmt.Sprintf("%d", snap.TotalErrors))
	var line string
	if m.width < 76 {
		// Compact header: drop workers/uptime so the row never wraps.
		line = fmt.Sprintf(" %s %s  hits %s  err %s",
			theme.Logo.Render("HITMAKER"), status, hits, errs)
	} else {
		line = fmt.Sprintf(" %s  %s  hits %s  errors %s  workers %d  uptime %s",
			theme.Logo.Render("HITMAKER"), status, hits, errs,
			snap.WorkerCount, snap.Uptime.Truncate(time.Second))
	}
	return lipgloss.NewStyle().Width(m.width).MaxHeight(2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Dim).Render(line)
}

func (m Model) tableView(snap simulator.Snapshot, width, height int) string {
	// All columns share one format after a 3-cell lead (space + glyph + space),
	// so the header and every row — selected or not — line up exactly. The URL
	// column takes whatever width is left; sizing it from `width` means a row can
	// never be wider than the pane and wrap.
	const cols = "%-22s %8s %8s %8s  %s"
	const lead = 3   // " " + glyph + " "
	const fixed = 51 // 22 + 1 + 8 + 1 + 8 + 1 + 8 + 2 (everything before URL)
	urlWidth := width - lead - fixed
	head := "   " + fmt.Sprintf(cols, "TARGET", "HITS", "RATE", "ERRORS", "URL")
	lines := []string{
		theme.Subtle.Render(head),
		theme.Subtle.Render(strings.Repeat("─", max(10, width))),
	}
	rows := max(1, height-len(lines))

	start := 0
	if m.selected >= rows {
		start = m.selected - rows + 1
	}
	end := min(len(snap.Targets), start+rows)
	for i := start; i < end; i++ {
		t := snap.Targets[i]
		name := trim(targetName(t.Target), 22)
		url := ""
		if urlWidth >= 6 {
			url = trim(t.Target, urlWidth)
		}
		cells := fmt.Sprintf(cols, name,
			fmt.Sprintf("%d", t.Hits), fmt.Sprintf("%d/m", t.CurrentRate), fmt.Sprintf("%d", t.Errors), url)
		if i == m.selected {
			// Quiet focus: an amber left tick, keeping the colored status glyph and
			// row text. No background fill, so the row can never wrap.
			lines = append(lines, theme.Tick.Render("▌")+statusGlyph(t)+" "+cells)
		} else {
			lines = append(lines, " "+statusGlyph(t)+" "+cells)
		}
	}
	if len(snap.Targets) == 0 {
		lines = append(lines, "", theme.Subtle.Render("  No targets yet."))
	}
	return clampLines(lines, height)
}

func statusGlyph(target simulator.TargetStats) string {
	switch {
	case target.Paused:
		return theme.Warn.Render("⏸")
	case target.Errors > 0 && target.Hits == 0:
		return theme.Bad.Render("✕")
	case target.ActiveWorkers == 0:
		return theme.Subtle.Render("○")
	default:
		return theme.Good.Render("●")
	}
}

func (m Model) recentView(snap simulator.Snapshot, width, height int) string {
	inner := max(12, width-4) // account for border (2) + padding (2)
	lines := []string{theme.Title.Render("Recent hits")}
	limit := max(0, height-3) // border (2) + title (1)
	for i, hit := range snap.Recent {
		if i >= limit {
			break
		}
		lines = append(lines, formatHit(hit, inner))
	}
	if len(snap.Recent) == 0 {
		lines = append(lines, theme.Subtle.Render("Waiting for the first hit…"))
	}
	body := padLines(lines, height-2)
	return theme.Border.Width(inner).Height(height - 2).Render(body)
}

func formatHit(hit simulator.HitResult, width int) string {
	status := theme.Good.Render(fmt.Sprintf("%d", hit.Status))
	if hit.Err != "" {
		status = theme.Bad.Render("ERR")
	}
	name := trim(targetName(hit.Target), max(4, width-16))
	return fmt.Sprintf("%s W%d %s %s", hit.At.Format("15:04:05"), hit.WorkerID, status, name)
}

// clampLines pads or truncates a slice of rendered lines to exactly n rows.
func clampLines(lines []string, n int) string {
	return padLines(lines, n)
}

func padLines(lines []string, n int) string {
	if n < 0 {
		n = 0
	}
	if len(lines) > n {
		lines = lines[:n]
	}
	for len(lines) < n {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func insetBlock(block string, left, top int) string {
	lines := strings.Split(block, "\n")
	prefix := strings.Repeat(" ", max(0, left))
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		} else if left > 0 {
			lines[i] = prefix
		}
	}
	if top > 0 {
		lines = append(make([]string, top), lines...)
	}
	return strings.Join(lines, "\n")
}

func (m Model) footerView() string {
	parts := []string{
		theme.Pill.Render("↑/↓ navigate"),
		theme.Pill.Render("K pause"),
		theme.Pill.Render("C config"),
		theme.Pill.Render("Q quit"),
	}
	return lipgloss.NewStyle().Width(m.width).PaddingTop(1).Render(strings.Join(parts, " "))
}

func tick() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func targetName(raw string) string {
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	return raw
}

func trim(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	if maxLen == 1 {
		return "…"
	}
	return value[:maxLen-1] + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
