package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zeb-link/hitmaker/internal/config"
)

// Guards the reported "settings won't stick": A then Enter in the standalone
// editor must persist to ./.hitmaker.json and quit.
func TestConfigModelApplySavesLocallyAndQuits(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Origin.Mode = config.ModeVercel
	var m tea.Model = NewConfigModel(cfg)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = m.Update(keyMsg("A")) // open apply preview
	_, cmd := m.Update(keyMsg("enter"))

	if cmd == nil {
		t.Fatal("expected a quit command after apply")
	}
	if msg := cmd(); msg == nil {
		t.Fatal("apply command produced no message")
	} else if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("apply should quit, got %T", msg)
	}

	saved, err := os.ReadFile(filepath.Join(dir, ".hitmaker.json"))
	if err != nil {
		t.Fatalf("local config not written: %v", err)
	}
	if !contains(string(saved), `"mode": "vercel"`) {
		t.Fatalf("saved config missing vercel mode:\n%s", saved)
	}
}
