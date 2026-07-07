package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootHelpSurfacesInteractiveCommands(t *testing.T) {
	opts := &rootOptions{Version: "test"}
	cmd := newRootCommand(opts)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"hitmaker tui links.txt", "hitmaker config edit", "run --for", "probe"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestConfigHelpIncludesEdit(t *testing.T) {
	opts := &rootOptions{Version: "test"}
	cmd := newRootCommand(opts)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"config", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "edit") {
		t.Fatalf("config help missing edit:\n%s", out.String())
	}
}

func TestExpandTargetsReadsFilesAndComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "links.txt")
	if err := os.WriteFile(path, []byte("\n# comment\nhttps://example.com/a\nhttps://example.com/b\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := expandTargets([]string{"https://example.com/direct", path})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://example.com/direct", "https://example.com/a", "https://example.com/b"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("targets = %#v, want %#v", got, want)
	}
}
