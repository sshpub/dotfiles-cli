package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExtractSection finds a section guard block in a module's shell files
// and returns the code body (without the guard line and closing brace).
//
// Section format: "module.section" (e.g. "git.shortcuts").
// Scans files listed in shell.load_order, or all .sh files if no load_order.
func ExtractSection(mod *Module, section string) (code string, sourceFile string, err error) {
	files := shellFiles(mod)
	if len(files) == 0 {
		return "", "", fmt.Errorf("module %q has no shell files", mod.Name)
	}

	guard := fmt.Sprintf(`dotfiles_section "%s"`, section)

	for _, file := range files {
		path := filepath.Join(mod.Dir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		body, found := extractGuardBlock(lines, guard)
		if found {
			return body, file, nil
		}
	}

	return "", "", fmt.Errorf("section %q not found in module %q shell files", section, mod.Name)
}

// extractGuardBlock finds a `dotfiles_section "X" && {` line and extracts
// the code body, tracking brace depth for nested blocks.
func extractGuardBlock(lines []string, guard string) (string, bool) {
	inBlock := false
	depth := 0
	var body []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inBlock {
			// Look for the guard line
			if strings.Contains(trimmed, guard) && strings.HasSuffix(trimmed, "{") {
				inBlock = true
				depth = 1
				continue
			}
			continue
		}

		// Count braces
		for _, ch := range line {
			switch ch {
			case '{':
				depth++
			case '}':
				depth--
			}
		}

		if depth <= 0 {
			// Closing brace of the guard block
			return strings.Join(body, "\n"), true
		}

		body = append(body, line)
	}

	return "", false
}

// WriteOverride writes extracted section code to an override file.
// The file includes a section guard so it integrates with the loader.
func WriteOverride(overrideDir string, section string, code string, sourceModule string, sourceFile string) error {
	if err := os.MkdirAll(overrideDir, 0755); err != nil {
		return err
	}

	fileName := section + ".sh"
	path := filepath.Join(overrideDir, fileName)

	var b strings.Builder
	fmt.Fprintf(&b, "#!/usr/bin/env bash\n")
	fmt.Fprintf(&b, "# Override: %s (extracted from modules/%s/%s)\n", section, sourceModule, sourceFile)
	fmt.Fprintf(&b, "# Edit this file to customize. Remove to restore module default.\n\n")
	fmt.Fprintf(&b, "dotfiles_section \"%s\" && {\n", section)
	b.WriteString(code)
	b.WriteString("\n}\n")

	return os.WriteFile(path, []byte(b.String()), 0644)
}

// ResetOverride removes override files for a section from both repo and local dirs.
func ResetOverride(repoOverrideDir string, localOverrideDir string, section string) (removed []string) {
	fileName := section + ".sh"

	for _, dir := range []string{repoOverrideDir, localOverrideDir} {
		if dir == "" {
			continue
		}
		path := filepath.Join(dir, fileName)
		if _, err := os.Stat(path); err == nil {
			os.Remove(path)
			removed = append(removed, path)
		}
	}

	return removed
}
