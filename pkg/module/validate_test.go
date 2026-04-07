package module

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSectionGuards_AllMatch(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "modules", "test")
	os.MkdirAll(modDir, 0755)

	os.WriteFile(filepath.Join(modDir, "aliases.sh"), []byte(`
dotfiles_section "test.main" && {
    alias foo="bar"
}
`), 0644)

	mod := &Module{
		Name:     "test",
		Dir:      modDir,
		Sections: map[string]string{"test.main": "Main"},
		Shell:    &ShellConfig{LoadOrder: []string{"aliases.sh"}},
	}

	warnings := ValidateSectionGuards(mod)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
}

func TestValidateSectionGuards_MissingGuard(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "modules", "test")
	os.MkdirAll(modDir, 0755)

	os.WriteFile(filepath.Join(modDir, "aliases.sh"), []byte(`
dotfiles_section "test.main" && {
    alias foo="bar"
}
`), 0644)

	mod := &Module{
		Name: "test",
		Dir:  modDir,
		Sections: map[string]string{
			"test.main":    "Main",
			"test.missing": "This has no guard",
		},
		Shell: &ShellConfig{LoadOrder: []string{"aliases.sh"}},
	}

	warnings := ValidateSectionGuards(mod)
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "test.missing") && strings.Contains(w, "no guard found") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning about test.missing, got: %v", warnings)
	}
}

func TestValidateSectionGuards_UndeclaredGuard(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "modules", "test")
	os.MkdirAll(modDir, 0755)

	os.WriteFile(filepath.Join(modDir, "aliases.sh"), []byte(`
dotfiles_section "test.main" && {
    alias foo="bar"
}
dotfiles_section "test.extra" && {
    alias baz="qux"
}
`), 0644)

	mod := &Module{
		Name:     "test",
		Dir:      modDir,
		Sections: map[string]string{"test.main": "Main"},
		Shell:    &ShellConfig{LoadOrder: []string{"aliases.sh"}},
	}

	warnings := ValidateSectionGuards(mod)
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "test.extra") && strings.Contains(w, "not declared") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning about test.extra, got: %v", warnings)
	}
}
