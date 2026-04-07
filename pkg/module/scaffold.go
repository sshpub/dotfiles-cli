package module

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Scaffold creates a new module directory with template files.
func Scaffold(dotfilesDir string, name string) (string, error) {
	if !namePattern.MatchString(name) {
		return "", fmt.Errorf("invalid module name %q: must match ^[a-z][a-z0-9-]*$", name)
	}

	modDir := filepath.Join(dotfilesDir, "modules", name)
	if _, err := os.Stat(modDir); err == nil {
		return "", fmt.Errorf("module directory already exists: %s", modDir)
	}

	if err := os.MkdirAll(modDir, 0755); err != nil {
		return "", err
	}

	// module.json
	mod := Module{
		Name:        name,
		Version:     "0.1.0",
		Description: "",
		Sections:    map[string]string{name + ".main": "Main section"},
		Shell:       &ShellConfig{LoadOrder: []string{"aliases.sh"}},
	}
	jsonData, err := json.MarshalIndent(mod, "", "  ")
	if err != nil {
		return "", err
	}
	jsonData = append(jsonData, '\n')
	if err := os.WriteFile(filepath.Join(modDir, "module.json"), jsonData, 0644); err != nil {
		return "", err
	}

	// aliases.sh
	shellContent := fmt.Sprintf("#!/usr/bin/env bash\n# modules/%s/aliases.sh\n\ndotfiles_section \"%s.main\" && {\n    # Add your aliases and functions here\n    :\n}\n", name, name)
	if err := os.WriteFile(filepath.Join(modDir, "aliases.sh"), []byte(shellContent), 0644); err != nil {
		return "", err
	}

	// CLAUDE.md
	claudeContent := fmt.Sprintf("# Module: %s\n\nDescribe this module for AI assistants.\n", name)
	if err := os.WriteFile(filepath.Join(modDir, "CLAUDE.md"), []byte(claudeContent), 0644); err != nil {
		return "", err
	}

	return modDir, nil
}
