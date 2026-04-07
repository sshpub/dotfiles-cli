package module

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testShellContent = `#!/usr/bin/env bash
# test module

dotfiles_section "test.aliases" && {
    alias foo="bar"
    alias baz="qux"
}

dotfiles_section "test.functions" && {
    myfunc() {
        if [[ -n "$1" ]]; then
            echo "$1"
        fi
    }
}
`

func setupTestModule(t *testing.T) *Module {
	t.Helper()
	dir := t.TempDir()
	modDir := filepath.Join(dir, "modules", "test")
	os.MkdirAll(modDir, 0755)
	os.WriteFile(filepath.Join(modDir, "aliases.sh"), []byte(testShellContent), 0644)
	os.WriteFile(filepath.Join(modDir, "module.json"), []byte(`{
		"name": "test", "version": "1.0.0", "description": "test",
		"sections": {"test.aliases": "Aliases", "test.functions": "Functions"},
		"shell": {"load_order": ["aliases.sh"]}
	}`), 0644)
	return &Module{
		Name: "test",
		Dir:  modDir,
		Shell: &ShellConfig{LoadOrder: []string{"aliases.sh"}},
		Sections: map[string]string{
			"test.aliases":   "Aliases",
			"test.functions": "Functions",
		},
	}
}

func TestExtractSection_Simple(t *testing.T) {
	mod := setupTestModule(t)
	code, file, err := ExtractSection(mod, "test.aliases")
	if err != nil {
		t.Fatalf("ExtractSection() error: %v", err)
	}
	if file != "aliases.sh" {
		t.Errorf("sourceFile = %q, want aliases.sh", file)
	}
	if !strings.Contains(code, `alias foo="bar"`) {
		t.Errorf("code missing foo alias:\n%s", code)
	}
	if !strings.Contains(code, `alias baz="qux"`) {
		t.Errorf("code missing baz alias:\n%s", code)
	}
}

func TestExtractSection_NestedBraces(t *testing.T) {
	mod := setupTestModule(t)
	code, _, err := ExtractSection(mod, "test.functions")
	if err != nil {
		t.Fatalf("ExtractSection() error: %v", err)
	}
	if !strings.Contains(code, "myfunc()") {
		t.Errorf("code missing function:\n%s", code)
	}
	if !strings.Contains(code, `echo "$1"`) {
		t.Errorf("code missing echo:\n%s", code)
	}
}

func TestExtractSection_NotFound(t *testing.T) {
	mod := setupTestModule(t)
	_, _, err := ExtractSection(mod, "test.nonexistent")
	if err == nil {
		t.Error("expected error for missing section")
	}
}

func TestWriteOverride(t *testing.T) {
	dir := t.TempDir()
	overrideDir := filepath.Join(dir, "overrides")

	code := "    alias foo=\"custom\""
	err := WriteOverride(overrideDir, "test.aliases", code, "test", "aliases.sh")
	if err != nil {
		t.Fatalf("WriteOverride() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(overrideDir, "test.aliases.sh"))
	if err != nil {
		t.Fatalf("reading override: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "# Override: test.aliases") {
		t.Error("missing override header")
	}
	if !strings.Contains(content, `dotfiles_section "test.aliases" && {`) {
		t.Error("missing section guard")
	}
	if !strings.Contains(content, `alias foo="custom"`) {
		t.Error("missing code body")
	}
}

func TestResetOverride(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "overrides")
	localDir := filepath.Join(dir, "local")
	os.MkdirAll(repoDir, 0755)
	os.MkdirAll(localDir, 0755)

	os.WriteFile(filepath.Join(repoDir, "test.aliases.sh"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(localDir, "test.aliases.sh"), []byte("test"), 0644)

	removed := ResetOverride(repoDir, localDir, "test.aliases")
	if len(removed) != 2 {
		t.Errorf("removed %d files, want 2", len(removed))
	}

	if _, err := os.Stat(filepath.Join(repoDir, "test.aliases.sh")); !os.IsNotExist(err) {
		t.Error("repo override should be removed")
	}
	if _, err := os.Stat(filepath.Join(localDir, "test.aliases.sh")); !os.IsNotExist(err) {
		t.Error("local override should be removed")
	}
}

func TestResetOverride_NoneExist(t *testing.T) {
	dir := t.TempDir()
	removed := ResetOverride(filepath.Join(dir, "overrides"), filepath.Join(dir, "local"), "test.aliases")
	if len(removed) != 0 {
		t.Errorf("removed %d files, want 0", len(removed))
	}
}
