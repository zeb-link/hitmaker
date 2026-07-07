package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kerns/hitmaker/internal/config"
)

type ConfigModel struct {
	editor  configEditor
	width   int
	height  int
	spinner spinner.Model
	err     error
}

func NewConfigModel(cfg config.Config) ConfigModel {
	spin := spinner.New()
	spin.Spinner = spinner.Spinner{Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}}
	editor := newConfigEditor(cfg)
	editor.status = "Edit, then press A to save & close. G/L save without closing, Esc discards."
	return ConfigModel{editor: editor, spinner: spin}
}

func (m ConfigModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		next, action, cmd := m.editor.Update(msg)
		m.editor = next
		switch action {
		case configActionClose:
			return m, tea.Quit
		case configActionApply:
			// Enter on the apply preview means "save & close": persist locally
			// by default, then exit. Any validation/write error keeps the editor
			// open so the user can see what went wrong.
			m.err = config.SaveLocal(m.editor.cfg)
			if m.err == nil {
				return m, tea.Quit
			}
		case configActionSaveGlobal:
			m.err = config.SaveGlobal(m.editor.cfg)
			if m.err == nil {
				m.editor.status = "Saved global config."
			}
		case configActionSaveLocal:
			m.err = config.SaveLocal(m.editor.cfg)
			if m.err == nil {
				m.editor.status = "Saved local .hitmaker.json."
			}
		case configActionDefaults:
			m.editor = newConfigEditor(config.Default())
			m.editor.status = "Defaults loaded in editor. Use G or L to save."
		}
		return m, cmd
	}
	return m, nil
}

func (m ConfigModel) View() string {
	width := m.width
	height := m.height
	if width == 0 {
		width = 100
	}
	if height == 0 {
		height = 32
	}
	return m.editor.View(width, height, m.err)
}
