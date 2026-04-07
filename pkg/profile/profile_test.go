package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDataDir_EnvOverride(t *testing.T) {
	t.Setenv("DOTFILES_DATA_DIR", "/tmp/test-data-dir")
	got := FindDataDir()
	if got != "/tmp/test-data-dir" {
		t.Errorf("FindDataDir() = %q, want /tmp/test-data-dir", got)
	}
}

func TestFindDataDir_Default(t *testing.T) {
	t.Setenv("DOTFILES_DATA_DIR", "")
	got := FindDataDir()
	if got == "" {
		t.Error("FindDataDir() returned empty string")
	}
}

func TestLoadProfile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")
	data := `{
		"role": ["personal"],
		"modules": {
			"git": true,
			"vim": {"shell": true, "disable": ["vim.plugins"]}
		},
		"git": {"name": "Test", "email": "test@example.com"},
		"modes": {
			"minimal": {
				"type": "include",
				"env_triggers": ["CI"],
				"include_modules": ["git"]
			}
		}
	}`
	os.WriteFile(path, []byte(data), 0644)

	p, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}
	if len(p.Role) != 1 || p.Role[0] != "personal" {
		t.Errorf("Role = %v, want [personal]", p.Role)
	}
	if p.Git == nil || p.Git.Name != "Test" {
		t.Error("Git config not loaded")
	}
	if len(p.Modes) != 1 {
		t.Errorf("Modes count = %d, want 1", len(p.Modes))
	}
}

func TestLoadProfile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("{invalid"), 0644)

	_, err := LoadProfile(path)
	if err == nil {
		t.Error("LoadProfile() expected error for invalid JSON")
	}
}

func TestLoadProfile_MissingFile(t *testing.T) {
	_, err := LoadProfile("/nonexistent/profile.json")
	if err == nil {
		t.Error("LoadProfile() expected error for missing file")
	}
}

func TestSaveProfile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	original := &Profile{
		Role: []string{"work"},
		Git:  &GitConfig{Name: "Test", Email: "test@example.com"},
		Modes: map[string]*Mode{
			"minimal": {Type: "include", EnvTriggers: []string{"CI"}},
		},
	}

	if err := SaveProfile(path, original); err != nil {
		t.Fatalf("SaveProfile() error: %v", err)
	}

	loaded, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}

	if len(loaded.Role) != 1 || loaded.Role[0] != "work" {
		t.Errorf("Role = %v, want [work]", loaded.Role)
	}
	if loaded.Git.Name != "Test" {
		t.Errorf("Git.Name = %q, want Test", loaded.Git.Name)
	}
	if loaded.Modes["minimal"].Type != "include" {
		t.Error("Mode type not preserved")
	}
}

func TestEnabledModules_Mixed(t *testing.T) {
	p := &Profile{
		Modules: map[string]interface{}{
			"git":    true,
			"vim":    map[string]interface{}{"shell": true},
			"node":   map[string]interface{}{"shell": false},
			"python": true,
		},
	}

	enabled := p.EnabledModules()
	want := []string{"git", "python", "vim"}
	if len(enabled) != len(want) {
		t.Fatalf("EnabledModules() = %v, want %v", enabled, want)
	}
	for i, name := range want {
		if enabled[i] != name {
			t.Errorf("EnabledModules()[%d] = %q, want %q", i, enabled[i], name)
		}
	}
}

func TestEnabledModules_ObjectNoShellField(t *testing.T) {
	// Object form without explicit "shell" field should count as enabled
	p := &Profile{
		Modules: map[string]interface{}{
			"git": map[string]interface{}{"install": true},
		},
	}
	enabled := p.EnabledModules()
	if len(enabled) != 1 || enabled[0] != "git" {
		t.Errorf("EnabledModules() = %v, want [git]", enabled)
	}
}

func TestEnabledModules_Nil(t *testing.T) {
	p := &Profile{}
	if got := p.EnabledModules(); got != nil {
		t.Errorf("EnabledModules() = %v, want nil", got)
	}
}

func TestDisabledSections(t *testing.T) {
	p := &Profile{
		Modules: map[string]interface{}{
			"git": true,
			"vim": map[string]interface{}{
				"shell":   true,
				"disable": []interface{}{"vim.plugins"},
			},
			"kubernetes": map[string]interface{}{
				"shell":   true,
				"disable": []interface{}{"kubernetes.helm", "kubernetes.istio"},
			},
		},
	}

	disabled := p.DisabledSections()
	want := []string{"kubernetes.helm", "kubernetes.istio", "vim.plugins"}
	if len(disabled) != len(want) {
		t.Fatalf("DisabledSections() = %v, want %v", disabled, want)
	}
	for i, sec := range want {
		if disabled[i] != sec {
			t.Errorf("DisabledSections()[%d] = %q, want %q", i, disabled[i], sec)
		}
	}
}

func TestFindDotfilesDir_EnvVar(t *testing.T) {
	t.Setenv("DOTFILES_DIR", "/tmp/dotfiles")
	got, err := FindDotfilesDir(nil)
	if err != nil {
		t.Fatalf("FindDotfilesDir() error: %v", err)
	}
	if got != "/tmp/dotfiles" {
		t.Errorf("FindDotfilesDir() = %q, want /tmp/dotfiles", got)
	}
}

func TestFindDotfilesDir_ProfileField(t *testing.T) {
	t.Setenv("DOTFILES_DIR", "")
	p := &Profile{DotfilesDir: "/home/user/dotfiles"}
	got, err := FindDotfilesDir(p)
	if err != nil {
		t.Fatalf("FindDotfilesDir() error: %v", err)
	}
	if got != "/home/user/dotfiles" {
		t.Errorf("FindDotfilesDir() = %q, want /home/user/dotfiles", got)
	}
}

func TestFindDotfilesDir_Error(t *testing.T) {
	t.Setenv("DOTFILES_DIR", "")
	_, err := FindDotfilesDir(&Profile{})
	if err == nil {
		t.Error("FindDotfilesDir() expected error when neither env nor profile set")
	}
}

func TestFindProfile_EnvVar(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-profile.json")
	os.WriteFile(path, []byte("{}"), 0644)

	t.Setenv("DOTFILES_PROFILE", path)
	got, err := FindProfile()
	if err != nil {
		t.Fatalf("FindProfile() error: %v", err)
	}
	if got != path {
		t.Errorf("FindProfile() = %q, want %q", got, path)
	}
}

func TestFindProfile_SearchChain(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOTFILES_PROFILE", "")
	t.Setenv("HOME", dir)
	t.Setenv("DOTFILES_DIR", "")

	// Create the second candidate in the chain
	configDir := filepath.Join(dir, ".config", "dotfiles")
	os.MkdirAll(configDir, 0755)
	path := filepath.Join(configDir, "profile.json")
	os.WriteFile(path, []byte("{}"), 0644)

	got, err := FindProfile()
	if err != nil {
		t.Fatalf("FindProfile() error: %v", err)
	}
	if got != path {
		t.Errorf("FindProfile() = %q, want %q", got, path)
	}
}

func TestFindProfile_NotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOTFILES_PROFILE", "")
	t.Setenv("HOME", dir)
	t.Setenv("DOTFILES_DIR", "")

	got, err := FindProfile()
	if err != nil {
		t.Fatalf("FindProfile() error: %v", err)
	}
	if got != "" {
		t.Errorf("FindProfile() = %q, want empty string", got)
	}
}
