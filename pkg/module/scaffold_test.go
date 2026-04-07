package module

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffold(t *testing.T) {
	dir := t.TempDir()
	modDir, err := Scaffold(dir, "my-tool")
	if err != nil {
		t.Fatalf("Scaffold() error: %v", err)
	}

	mod, err := LoadModule(filepath.Join(modDir, "module.json"))
	if err != nil {
		t.Fatalf("LoadModule() error: %v", err)
	}
	if mod.Name != "my-tool" {
		t.Errorf("name = %q, want my-tool", mod.Name)
	}
	if mod.Version != "0.1.0" {
		t.Errorf("version = %q, want 0.1.0", mod.Version)
	}
	if mod.Sections["my-tool.main"] != "Main section" {
		t.Error("missing my-tool.main section")
	}

	data, err := os.ReadFile(filepath.Join(modDir, "aliases.sh"))
	if err != nil {
		t.Fatalf("reading aliases.sh: %v", err)
	}
	if !strings.Contains(string(data), `dotfiles_section "my-tool.main"`) {
		t.Error("aliases.sh missing section guard")
	}

	if _, err := os.Stat(filepath.Join(modDir, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not created")
	}
}

func TestScaffold_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "modules", "existing"), 0755)
	_, err := Scaffold(dir, "existing")
	if err == nil {
		t.Error("expected error for existing directory")
	}
}

func TestScaffold_InvalidName(t *testing.T) {
	dir := t.TempDir()
	_, err := Scaffold(dir, "Bad_Name")
	if err == nil {
		t.Error("expected error for invalid name")
	}
}
