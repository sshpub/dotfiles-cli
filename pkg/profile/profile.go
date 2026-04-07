package profile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

// Profile mirrors the ~/.dotfiles.json schema.
type Profile struct {
	Comment     string                 `json:"_comment,omitempty"`
	Role        []string               `json:"role,omitempty"`
	Platform    *PlatformOverride      `json:"platform,omitempty"`
	Modules     map[string]interface{} `json:"modules,omitempty"`
	Git         *GitConfig             `json:"git,omitempty"`
	Modes       map[string]*Mode       `json:"modes,omitempty"`
	DotfilesDir string                 `json:"dotfiles_dir,omitempty"`
	Registries  []Registry             `json:"registries,omitempty"`
}

type PlatformOverride struct {
	Comment string `json:"_comment,omitempty"`
	OS      string `json:"os,omitempty"`
	Variant string `json:"variant,omitempty"`
	Distro  string `json:"distro,omitempty"`
}

type GitConfig struct {
	Comment string `json:"_comment,omitempty"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
}

type Mode struct {
	Comment        string   `json:"_comment,omitempty"`
	Type           string   `json:"type,omitempty"`
	EnvTriggers    []string `json:"env_triggers,omitempty"`
	IncludeModules []string `json:"include_modules,omitempty"`
	NeverLoad      []string `json:"never_load,omitempty"`
}

type Registry struct {
	Comment string `json:"_comment,omitempty"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Private bool   `json:"private,omitempty"`
}

// FindProfile searches the standard chain and returns the path to the first
// profile found. Returns empty string and nil if no profile exists.
func FindProfile() (string, error) {
	if p := os.Getenv("DOTFILES_PROFILE"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dotfilesDir := os.Getenv("DOTFILES_DIR")

	candidates := []string{
		filepath.Join(home, ".dotfiles.json"),
		filepath.Join(home, ".config", "dotfiles", "profile.json"),
		filepath.Join(home, ".config", "dotfiles.json"),
		filepath.Join(home, ".local", "dotfiles.json"),
	}

	if dotfilesDir != "" {
		candidates = append(candidates,
			filepath.Join(dotfilesDir, "dotfiles.json"),
			filepath.Join(dotfilesDir, "profiles", "default.json"),
		)
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", nil
}

// LoadProfile parses a JSON profile file.
func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// SaveProfile writes a profile as pretty-printed JSON.
func SaveProfile(path string, p *Profile) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// FindDotfilesDir resolves the dotfiles repo directory.
// Search: $DOTFILES_DIR env > profile's dotfiles_dir field > error.
func FindDotfilesDir(p *Profile) (string, error) {
	if d := os.Getenv("DOTFILES_DIR"); d != "" {
		return d, nil
	}
	if p != nil && p.DotfilesDir != "" {
		return p.DotfilesDir, nil
	}
	return "", errors.New("dotfiles directory not found: set $DOTFILES_DIR or dotfiles_dir in profile")
}

// FindDataDir resolves the data directory.
// Search: $DOTFILES_DATA_DIR > first existing of ~/.dotfiles,
// ~/.config/dotfiles, ~/.local/share/dotfiles > default ~/.dotfiles.
func FindDataDir() string {
	if d := os.Getenv("DOTFILES_DATA_DIR"); d != "" {
		return d
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".dotfiles")
	}

	candidates := []string{
		filepath.Join(home, ".dotfiles"),
		filepath.Join(home, ".config", "dotfiles"),
		filepath.Join(home, ".local", "share", "dotfiles"),
	}

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}

	return filepath.Join(home, ".dotfiles")
}

// EnabledModules returns the sorted list of module names that have shell
// loading enabled. A module entry of `true` (bool) or `{"shell": true}` counts.
func (p *Profile) EnabledModules() []string {
	if p.Modules == nil {
		return nil
	}
	var enabled []string
	for name, val := range p.Modules {
		switch v := val.(type) {
		case bool:
			if v {
				enabled = append(enabled, name)
			}
		case map[string]interface{}:
			shell, ok := v["shell"]
			if !ok || shell == true {
				enabled = append(enabled, name)
			}
		}
	}
	sort.Strings(enabled)
	return enabled
}

// DisabledSections returns all disabled sections across all modules, sorted.
func (p *Profile) DisabledSections() []string {
	if p.Modules == nil {
		return nil
	}
	var disabled []string
	for _, val := range p.Modules {
		obj, ok := val.(map[string]interface{})
		if !ok {
			continue
		}
		disableRaw, ok := obj["disable"]
		if !ok {
			continue
		}
		arr, ok := disableRaw.([]interface{})
		if !ok {
			continue
		}
		for _, item := range arr {
			if s, ok := item.(string); ok {
				disabled = append(disabled, s)
			}
		}
	}
	sort.Strings(disabled)
	return disabled
}
