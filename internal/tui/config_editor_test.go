package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kerns/hitmaker/internal/config"
)

func TestConfigEditorAdvertisedUppercaseShortcuts(t *testing.T) {
	tests := []struct {
		key  string
		want configAction
	}{
		{key: "G", want: configActionSaveGlobal},
		{key: "L", want: configActionSaveLocal},
		{key: "D", want: configActionDefaults},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			editor := newConfigEditor(config.Default())
			_, got, _ := editor.Update(keyMsg(tt.key))
			if got != tt.want {
				t.Fatalf("action = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigEditorUppercaseNestedParamKeys(t *testing.T) {
	editor := newConfigEditor(config.Default())
	editor.focus = len(editor.fields) - 1
	editor, _, _ = editor.Update(keyMsg("enter"))
	if editor.pane != paneParams {
		t.Fatalf("pane = %v, want params", editor.pane)
	}
	before := len(editor.cfg.Requests.URLParams)
	editor, _, _ = editor.Update(keyMsg("N"))
	if got := len(editor.cfg.Requests.URLParams); got != before+1 {
		t.Fatalf("params len = %d, want %d", got, before+1)
	}
	editor, _, _ = editor.Update(keyMsg("P"))
	if editor.pane != panePayloads {
		t.Fatalf("pane = %v, want payloads", editor.pane)
	}
}

func TestConfigEditorViewShowsCommandBarAndSections(t *testing.T) {
	editor := newConfigEditor(config.Default())
	view := editor.View(120, 36, nil)
	for _, want := range []string{"TRAFFIC", "IDENTITY", "ORIGIN", "SHORTCUTS", "URL PARAMS", "Type numbers"} {
		if !contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if contains(view, "LIVE PREVIEW") {
		t.Fatalf("preview should not be pinned on the main editor:\n%s", view)
	}
}

func TestConfigEditorDirectNumberTyping(t *testing.T) {
	editor := newConfigEditor(config.Default())
	editor, action, _ := editor.Update(keyMsg("4"))
	if action != configActionNone {
		t.Fatalf("action = %v, want none", action)
	}
	if editor.cfg.Traffic.MinPerMin != 4 {
		t.Fatalf("min rate = %d, want 4", editor.cfg.Traffic.MinPerMin)
	}
	editor, _, _ = editor.Update(keyMsg("2"))
	if editor.cfg.Traffic.MinPerMin != 42 {
		t.Fatalf("min rate = %d, want 42", editor.cfg.Traffic.MinPerMin)
	}
}

func TestConfigEditorApplyShowsPreviewBeforeAction(t *testing.T) {
	editor := newConfigEditor(config.Default())
	editor, action, _ := editor.Update(keyMsg("A"))
	if action != configActionNone {
		t.Fatalf("initial apply action = %v, want none", action)
	}
	if editor.pane != paneConfirmApply {
		t.Fatalf("pane = %v, want confirm apply", editor.pane)
	}
	view := editor.View(100, 30, nil)
	if !contains(view, "SAVE & CLOSE") || !contains(view, "TRAFFIC") {
		t.Fatalf("apply preview missing expected content:\n%s", view)
	}
	editor, action, _ = editor.Update(keyMsg("enter"))
	if action != configActionApply {
		t.Fatalf("confirm action = %v, want apply", action)
	}
}

func TestConfigEditorDimsProxyFieldsWhenNotProxyMode(t *testing.T) {
	editor := newConfigEditor(config.Default())
	editor.cfg.Origin.Mode = config.ModeNone
	for i, field := range editor.fields {
		if field.key == "provider" {
			editor.focus = i
			break
		}
	}
	view := editor.View(120, 36, nil)
	if !contains(view, "Disabled until Origin mode is Proxy service.") {
		t.Fatalf("proxy disabled help missing:\n%s", view)
	}
}

func TestConfigEditorFitsStandardTerminal(t *testing.T) {
	editor := newConfigEditor(config.Default())
	view := editor.View(80, 24, nil)
	if got := lipgloss.Height(view); got > 24 {
		t.Fatalf("view height = %d, want <= 24:\n%s", got, view)
	}
	for _, want := range []string{"CONTROL DECK", "KEYS"} {
		if !contains(view, want) {
			t.Fatalf("compact view missing %q:\n%s", want, view)
		}
	}
}

func keyMsg(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func contains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
