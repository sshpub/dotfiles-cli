package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePlatformCache(t *testing.T) {
	dir := t.TempDir()
	info := &PlatformInfo{
		OS:             "linux",
		Arch:           "amd64",
		WSL:            true,
		Distro:         "ubuntu",
		DistroVersion:  "25.10",
		PkgManager:     "apt",
		Container:      false,
		HomebrewPrefix: "",
		MacOSVersion:   "",
	}

	if err := GeneratePlatformCache(dir, info); err != nil {
		t.Fatalf("GeneratePlatformCache() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "cache", "platform.sh"))
	if err != nil {
		t.Fatalf("reading cache file: %v", err)
	}

	content := string(data)
	checks := map[string]string{
		"DOTFILES_OS":             `DOTFILES_OS="linux"`,
		"DOTFILES_ARCH":           `DOTFILES_ARCH="amd64"`,
		"DOTFILES_WSL":            `DOTFILES_WSL="true"`,
		"DOTFILES_DISTRO":         `DOTFILES_DISTRO="ubuntu"`,
		"DOTFILES_DISTRO_VERSION": `DOTFILES_DISTRO_VERSION="25.10"`,
		"DOTFILES_PKG_MANAGER":    `DOTFILES_PKG_MANAGER="apt"`,
		"DOTFILES_CONTAINER":      `DOTFILES_CONTAINER=""`,
		"HOMEBREW_PREFIX":         `HOMEBREW_PREFIX=""`,
		"MACOS_VERSION":           `MACOS_VERSION=""`,
	}

	for name, expected := range checks {
		if !strings.Contains(content, expected) {
			t.Errorf("%s: expected %q in output, got:\n%s", name, expected, content)
		}
	}
}

func TestGeneratePlatformCache_Container(t *testing.T) {
	dir := t.TempDir()
	info := &PlatformInfo{
		OS:        "linux",
		Arch:      "arm64",
		Container: true,
	}

	if err := GeneratePlatformCache(dir, info); err != nil {
		t.Fatalf("GeneratePlatformCache() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "cache", "platform.sh"))
	if !strings.Contains(string(data), `DOTFILES_CONTAINER="true"`) {
		t.Error("expected DOTFILES_CONTAINER=\"true\"")
	}
}

func TestGenerateProfileCache(t *testing.T) {
	dir := t.TempDir()
	p := &Profile{
		Modules: map[string]interface{}{
			"git":    true,
			"vim":    map[string]interface{}{"shell": true, "disable": []interface{}{"vim.plugins"}},
			"python": true,
		},
		Modes: map[string]*Mode{
			"minimal": {
				Type:           "include",
				EnvTriggers:    []string{"CI", "CLAUDE_CODE"},
				IncludeModules: []string{"git"},
			},
			"server": {
				Type:      "exclude",
				NeverLoad: []string{"vim", "python"},
			},
		},
	}

	profilePath := "/home/user/.dotfiles.json"
	if err := GenerateProfileCache(dir, profilePath, p); err != nil {
		t.Fatalf("GenerateProfileCache() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "cache", "profile.sh"))
	if err != nil {
		t.Fatalf("reading cache file: %v", err)
	}

	content := string(data)

	// Check profile source
	if !strings.Contains(content, `DOTFILES_PROFILE_SOURCE="/home/user/.dotfiles.json"`) {
		t.Error("missing DOTFILES_PROFILE_SOURCE")
	}

	// Check enabled modules (sorted)
	if !strings.Contains(content, "DOTFILES_ENABLED_MODULES=(git python vim)") {
		t.Errorf("unexpected enabled modules in:\n%s", content)
	}

	// Check disabled sections
	if !strings.Contains(content, "DOTFILES_DISABLED_SECTIONS=(vim.plugins)") {
		t.Errorf("unexpected disabled sections in:\n%s", content)
	}

	// Check mode names (sorted)
	if !strings.Contains(content, "DOTFILES_MODE_NAMES=(minimal server)") {
		t.Errorf("unexpected mode names in:\n%s", content)
	}

	// Check minimal mode details
	if !strings.Contains(content, "DOTFILES_MODE_minimal_TYPE=include") {
		t.Error("missing minimal mode type")
	}
	if !strings.Contains(content, "DOTFILES_MODE_minimal_TRIGGERS=(CI CLAUDE_CODE)") {
		t.Error("missing minimal mode triggers")
	}
	if !strings.Contains(content, "DOTFILES_MODE_minimal_MODULES=(git)") {
		t.Error("missing minimal mode modules")
	}

	// Check server mode details
	if !strings.Contains(content, "DOTFILES_MODE_server_TYPE=exclude") {
		t.Error("missing server mode type")
	}
	if !strings.Contains(content, "DOTFILES_MODE_server_NEVER_LOAD=(vim python)") {
		t.Error("missing server mode never_load")
	}
}

func TestGenerateProfileCache_Empty(t *testing.T) {
	dir := t.TempDir()
	p := &Profile{}

	if err := GenerateProfileCache(dir, "/empty.json", p); err != nil {
		t.Fatalf("GenerateProfileCache() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "cache", "profile.sh"))
	content := string(data)

	if !strings.Contains(content, "DOTFILES_ENABLED_MODULES=()") {
		t.Error("empty profile should produce empty DOTFILES_ENABLED_MODULES=()")
	}
	if !strings.Contains(content, "DOTFILES_DISABLED_SECTIONS=()") {
		t.Error("empty profile should produce empty DOTFILES_DISABLED_SECTIONS=()")
	}
	if !strings.Contains(content, "DOTFILES_MODE_NAMES=()") {
		t.Error("empty profile should produce empty DOTFILES_MODE_NAMES=()")
	}
}

func TestClearCache(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(filepath.Join(cacheDir, "platform.sh"), []byte("test"), 0644)

	if err := ClearCache(dir); err != nil {
		t.Fatalf("ClearCache() error: %v", err)
	}

	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Error("cache directory should be removed")
	}
}

func TestClearCache_NonExistent(t *testing.T) {
	dir := t.TempDir()
	// No cache dir exists — should not error
	if err := ClearCache(dir); err != nil {
		t.Fatalf("ClearCache() error on non-existent dir: %v", err)
	}
}

func TestGeneratePlatformCache_BashSyntax(t *testing.T) {
	dir := t.TempDir()
	info := &PlatformInfo{
		OS:   "linux",
		Arch: "amd64",
	}

	if err := GeneratePlatformCache(dir, info); err != nil {
		t.Fatalf("GeneratePlatformCache() error: %v", err)
	}

	// Verify the output is valid bash
	cacheFile := filepath.Join(dir, "cache", "platform.sh")
	out, err := execCommand("bash", "-n", cacheFile)
	if err != nil {
		t.Errorf("platform cache is not valid bash: %v\n%s", err, out)
	}
}
