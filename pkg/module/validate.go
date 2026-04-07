package module

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var sectionGuardRe = regexp.MustCompile(`dotfiles_section\s+"([^"]+)"`)

// ValidateSectionGuards compares declared sections in module.json against
// actual dotfiles_section calls in the shell files.
func ValidateSectionGuards(mod *Module) []string {
	var warnings []string

	foundGuards := make(map[string]bool)
	files := shellFiles(mod)

	for _, file := range files {
		path := filepath.Join(mod.Dir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("cannot read %s: %v", file, err))
			continue
		}

		matches := sectionGuardRe.FindAllStringSubmatch(string(data), -1)
		for _, match := range matches {
			foundGuards[match[1]] = true
		}
	}

	for section := range mod.Sections {
		if !foundGuards[section] {
			warnings = append(warnings, fmt.Sprintf("section %q declared in module.json but no guard found in shell files", section))
		}
	}

	for guard := range foundGuards {
		if !strings.HasPrefix(guard, mod.Name+".") {
			continue
		}
		if _, declared := mod.Sections[guard]; !declared {
			warnings = append(warnings, fmt.Sprintf("guard %q found in shell files but not declared in module.json sections", guard))
		}
	}

	return warnings
}

// shellFiles returns the list of .sh files for a module.
func shellFiles(mod *Module) []string {
	if mod.Shell != nil && len(mod.Shell.LoadOrder) > 0 {
		return mod.Shell.LoadOrder
	}

	entries, err := os.ReadDir(mod.Dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sh") {
			files = append(files, e.Name())
		}
	}
	return files
}
