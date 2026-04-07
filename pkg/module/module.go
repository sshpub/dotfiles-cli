package module

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Module mirrors the modules/schema.json definition.
type Module struct {
	Comment      string                     `json:"_comment,omitempty"`
	Name         string                     `json:"name"`
	Version      string                     `json:"version"`
	Description  string                     `json:"description"`
	Author       string                     `json:"author,omitempty"`
	Platforms    []string                   `json:"platforms,omitempty"`
	Dependencies []string                   `json:"dependencies,omitempty"`
	Sections     map[string]string          `json:"sections,omitempty"`
	Shell        *ShellConfig               `json:"shell,omitempty"`
	Install      map[string]*InstallRecipes `json:"install,omitempty"`
	Symlinks     map[string]string          `json:"symlinks,omitempty"`
	Hooks        *HookConfig                `json:"hooks,omitempty"`

	// Dir is the absolute path to the module directory (set by DiscoverModules).
	Dir string `json:"-"`
}

type ShellConfig struct {
	LoadOrder []string `json:"load_order,omitempty"`
}

type InstallRecipes struct {
	Brew    []string `json:"brew,omitempty"`
	Apt     []string `json:"apt,omitempty"`
	Dnf     []string `json:"dnf,omitempty"`
	Pacman  []string `json:"pacman,omitempty"`
	Snap    []string `json:"snap,omitempty"`
	Zypper  []string `json:"zypper,omitempty"`
	Inherit string   `json:"inherit,omitempty"`
}

type HookConfig struct {
	PostInstall string `json:"post_install,omitempty"`
	PostEnable  string `json:"post_enable,omitempty"`
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// DiscoverModules finds all modules/*/module.json under dotfilesDir.
func DiscoverModules(dotfilesDir string) ([]Module, error) {
	modulesDir := filepath.Join(dotfilesDir, "modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return nil, fmt.Errorf("reading modules directory: %w", err)
	}

	var modules []Module
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		jsonPath := filepath.Join(modulesDir, entry.Name(), "module.json")
		if _, err := os.Stat(jsonPath); err != nil {
			continue
		}
		mod, err := LoadModule(jsonPath)
		if err != nil {
			continue // skip invalid modules
		}
		mod.Dir = filepath.Join(modulesDir, entry.Name())
		modules = append(modules, *mod)
	}

	return modules, nil
}

// LoadModule parses a single module.json file.
func LoadModule(path string) (*Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Module
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ValidateModule checks a module and returns a list of errors.
// Returns nil if valid.
func ValidateModule(mod *Module) []string {
	var errs []string

	if mod.Name == "" {
		errs = append(errs, "name is required")
	} else if !namePattern.MatchString(mod.Name) {
		errs = append(errs, fmt.Sprintf("name %q must match ^[a-z][a-z0-9-]*$", mod.Name))
	}

	if mod.Version == "" {
		errs = append(errs, "version is required")
	} else if !versionPattern.MatchString(mod.Version) {
		errs = append(errs, fmt.Sprintf("version %q must match semver (e.g. 1.0.0)", mod.Version))
	}

	if mod.Description == "" {
		errs = append(errs, "description is required")
	}

	for _, dep := range mod.Dependencies {
		if !namePattern.MatchString(dep) {
			errs = append(errs, fmt.Sprintf("dependency %q is not a valid module name", dep))
		}
	}

	for _, p := range mod.Platforms {
		switch p {
		case "macos", "linux", "wsl":
			// valid
		default:
			errs = append(errs, fmt.Sprintf("platform %q must be macos, linux, or wsl", p))
		}
	}

	return errs
}
