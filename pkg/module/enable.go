package module

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Enable sets a module to enabled (true) in the profile's modules map,
// saves the profile, and returns the updated profile bytes.
// Verifies the module directory exists under dotfilesDir.
func Enable(profilePath string, moduleName string, dotfilesDir string) error {
	// Verify module exists on disk
	modDir := filepath.Join(dotfilesDir, "modules", moduleName)
	jsonPath := filepath.Join(modDir, "module.json")
	if _, err := os.Stat(jsonPath); err != nil {
		return fmt.Errorf("module %q not found at %s", moduleName, modDir)
	}

	return mutateProfile(profilePath, func(data map[string]interface{}) {
		modules, ok := data["modules"].(map[string]interface{})
		if !ok {
			modules = make(map[string]interface{})
			data["modules"] = modules
		}
		modules[moduleName] = true
	})
}

// Disable sets a module to disabled (false) in the profile's modules map
// and saves the profile. Does not require the module directory to exist.
func Disable(profilePath string, moduleName string) error {
	return mutateProfile(profilePath, func(data map[string]interface{}) {
		modules, ok := data["modules"].(map[string]interface{})
		if !ok {
			modules = make(map[string]interface{})
			data["modules"] = modules
		}
		modules[moduleName] = false
	})
}

// DisableSection adds a section to the module's disable list in the profile.
func DisableSection(profilePath string, moduleName string, section string) error {
	return mutateProfile(profilePath, func(data map[string]interface{}) {
		modules, ok := data["modules"].(map[string]interface{})
		if !ok {
			modules = make(map[string]interface{})
			data["modules"] = modules
		}

		// Get or create the module object
		modRaw, exists := modules[moduleName]
		var modObj map[string]interface{}

		switch v := modRaw.(type) {
		case map[string]interface{}:
			modObj = v
		case bool:
			modObj = map[string]interface{}{"shell": v}
		default:
			if !exists {
				modObj = map[string]interface{}{"shell": true}
			} else {
				modObj = map[string]interface{}{}
			}
		}

		// Get or create disable list
		disableRaw, _ := modObj["disable"].([]interface{})
		for _, d := range disableRaw {
			if s, ok := d.(string); ok && s == section {
				return // already disabled
			}
		}
		modObj["disable"] = append(disableRaw, section)
		modules[moduleName] = modObj
	})
}

// EnableSection removes a section from the module's disable list in the profile.
func EnableSection(profilePath string, moduleName string, section string) error {
	return mutateProfile(profilePath, func(data map[string]interface{}) {
		modules, ok := data["modules"].(map[string]interface{})
		if !ok {
			return
		}
		modObj, ok := modules[moduleName].(map[string]interface{})
		if !ok {
			return
		}
		disableRaw, ok := modObj["disable"].([]interface{})
		if !ok {
			return
		}
		filtered := make([]interface{}, 0, len(disableRaw))
		for _, d := range disableRaw {
			if s, ok := d.(string); ok && s == section {
				continue
			}
			filtered = append(filtered, d)
		}
		if len(filtered) == 0 {
			delete(modObj, "disable")
		} else {
			modObj["disable"] = filtered
		}
	})
}

// mutateProfile reads a profile JSON, applies a mutation, and writes it back.
// Preserves all fields including unknown ones (no struct round-trip).
func mutateProfile(profilePath string, fn func(data map[string]interface{})) error {
	raw, err := os.ReadFile(profilePath)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	fn(data)

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(profilePath, out, 0644)
}
