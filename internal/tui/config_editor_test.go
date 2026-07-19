package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/zeb-link/hitmaker/v2/internal/config"
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
	// Tall enough that the whole control deck fits without scrolling — the deck
	// grew an ENTROPY group, so URL PARAMS (the last row) needs the extra height.
	view := editor.View(120, 60, nil)
	for _, want := range []string{"TRAFFIC", "IDENTITY", "SCHEDULE", "ENTROPY", "ORIGIN", "URL PARAMS", "save & close", "move"} {
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

func TestConfigEditorSelectChangesInlineAndEnterAdvances(t *testing.T) {
	editor := newConfigEditor(config.Default())
	methodFocus := 0
	for i, field := range editor.fields {
		if field.key == "method" {
			methodFocus = i
			editor.focus = i
			break
		}
	}

	editor, action, _ := editor.Update(keyMsg("n"))
	if action != configActionNone {
		t.Fatalf("action = %v, want none", action)
	}
	if editor.cfg.Requests.Method != "HEAD" {
		t.Fatalf("method = %s, want HEAD", editor.cfg.Requests.Method)
	}
	if editor.focus != methodFocus {
		t.Fatalf("focus = %d, want method row %d", editor.focus, methodFocus)
	}

	editor, _, _ = editor.Update(keyMsg("j"))
	if editor.focus != methodFocus+1 {
		t.Fatalf("down focus = %d, want next row %d", editor.focus, methodFocus+1)
	}
	editor.focus = methodFocus

	editor, _, _ = editor.Update(keyMsg("enter"))
	if editor.focus != methodFocus+1 {
		t.Fatalf("enter focus = %d, want next row %d", editor.focus, methodFocus+1)
	}
}

func TestConfigEditorFocusedSelectKeepsRadioRendering(t *testing.T) {
	editor := newConfigEditor(config.Default())
	for i, field := range editor.fields {
		if field.key == "method" {
			editor.focus = i
			break
		}
	}
	view := editor.View(120, 36, nil)
	if contains(view, "[GET]") {
		t.Fatalf("focused select should not use bracket fallback:\n%s", view)
	}
	if !contains(view, "● GET") || !contains(view, "○ HEAD") {
		t.Fatalf("focused select missing radio options:\n%s", view)
	}
}

func TestConfigEditorWideViewShowsFieldGuide(t *testing.T) {
	editor := newConfigEditor(config.Default())
	for i, field := range editor.fields {
		if field.key == "mode" {
			editor.focus = i
			break
		}
	}
	view := editor.View(160, 36, nil)
	for _, want := range []string{"FIELD GUIDE", "ORIGIN / Origin mode", "Auto:", "public domains", "localhost"} {
		if !contains(view, want) {
			t.Fatalf("wide view missing %q:\n%s", want, view)
		}
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
	if !contains(view, "Apply changes") || !contains(view, "hits/min") {
		t.Fatalf("apply modal missing expected content:\n%s", view)
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
	if !contains(view, "Enable Auto or Proxy service") {
		t.Fatalf("proxy disabled help missing:\n%s", view)
	}
}

func TestConfigEditorFitsStandardTerminal(t *testing.T) {
	editor := newConfigEditor(config.Default())
	view := editor.View(80, 24, nil)
	if got := lipgloss.Height(view); got > 24 {
		t.Fatalf("view height = %d, want <= 24:\n%s", got, view)
	}
	for _, want := range []string{"CONTROL DECK", "move"} {
		if !contains(view, want) {
			t.Fatalf("compact view missing %q:\n%s", want, view)
		}
	}
}

func TestFieldGuideWrapsWithoutOrphans(t *testing.T) {
	editor := newConfigEditor(config.Default())
	for i, f := range editor.fields {
		if f.key == "mode" {
			editor.focus = i
			break
		}
	}
	out := stripANSI(editor.fieldGuideView(64, 40))
	// The double-wrap bug split trailing words onto their own line. The summary
	// should keep words together up to the real text width.
	if !contains(out, "come from. Identity,") {
		t.Fatalf("summary double-wrapped — expected 'come from. Identity,' on one line:\n%s", out)
	}
	for _, ln := range strings.Split(out, "\n") {
		word := strings.TrimSpace(strings.Trim(ln, "│╭╮╰╯─ "))
		if word == "Identity," || word == "internal" {
			t.Fatalf("orphaned word %q — field guide is double-wrapping:\n%s", word, out)
		}
	}
}

func keyMsg(value string) tea.KeyPressMsg {
	switch value {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	default:
		r := []rune(value)
		return tea.KeyPressMsg{Code: r[0], Text: value}
	}
}

func contains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
