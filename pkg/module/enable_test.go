package module

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTestProfile(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "profile.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test profile: %v", err)
	}
	return path
}

func readProfileModules(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading profile: %v", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("parsing profile: %v", err)
	}
	modules, _ := data["modules"].(map[string]interface{})
	return modules
}

func TestEnable_NewModule(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"modules": {"git": true}}`)

	modDir := filepath.Join(dir, "modules", "vim")
	os.MkdirAll(modDir, 0755)
	os.WriteFile(filepath.Join(modDir, "module.json"), []byte(`{"name":"vim","version":"1.0.0","description":"test"}`), 0644)

	err := Enable(profilePath, "vim", dir)
	if err != nil {
		t.Fatalf("Enable() error: %v", err)
	}

	modules := readProfileModules(t, profilePath)
	if modules["vim"] != true {
		t.Errorf("vim = %v, want true", modules["vim"])
	}
	if modules["git"] != true {
		t.Errorf("git = %v, want true (preserved)", modules["git"])
	}
}

func TestEnable_ModuleNotFound(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"modules": {}}`)

	err := Enable(profilePath, "nonexistent", dir)
	if err == nil {
		t.Error("Enable() expected error for missing module")
	}
}

func TestDisable(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"modules": {"git": true, "vim": true}}`)

	err := Disable(profilePath, "vim")
	if err != nil {
		t.Fatalf("Disable() error: %v", err)
	}

	modules := readProfileModules(t, profilePath)
	if modules["vim"] != false {
		t.Errorf("vim = %v, want false", modules["vim"])
	}
	if modules["git"] != true {
		t.Errorf("git = %v, want true (preserved)", modules["git"])
	}
}

func TestDisable_NoModulesMap(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"role": ["personal"]}`)

	err := Disable(profilePath, "vim")
	if err != nil {
		t.Fatalf("Disable() error: %v", err)
	}

	modules := readProfileModules(t, profilePath)
	if modules["vim"] != false {
		t.Errorf("vim = %v, want false", modules["vim"])
	}
}

func TestDisableSection(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"modules": {"git": true}}`)

	err := DisableSection(profilePath, "git", "git.log")
	if err != nil {
		t.Fatalf("DisableSection() error: %v", err)
	}

	modules := readProfileModules(t, profilePath)
	gitObj, ok := modules["git"].(map[string]interface{})
	if !ok {
		t.Fatalf("git should be object, got %T", modules["git"])
	}
	disableRaw, _ := gitObj["disable"].([]interface{})
	if len(disableRaw) != 1 || disableRaw[0] != "git.log" {
		t.Errorf("disable = %v, want [git.log]", disableRaw)
	}
	if gitObj["shell"] != true {
		t.Errorf("shell = %v, want true", gitObj["shell"])
	}
}

func TestDisableSection_NoDuplicate(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"modules": {"git": {"shell": true, "disable": ["git.log"]}}}`)

	err := DisableSection(profilePath, "git", "git.log")
	if err != nil {
		t.Fatalf("DisableSection() error: %v", err)
	}

	modules := readProfileModules(t, profilePath)
	gitObj := modules["git"].(map[string]interface{})
	disableRaw := gitObj["disable"].([]interface{})
	if len(disableRaw) != 1 {
		t.Errorf("disable = %v, should not duplicate", disableRaw)
	}
}

func TestEnableSection(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"modules": {"git": {"shell": true, "disable": ["git.log", "git.branch"]}}}`)

	err := EnableSection(profilePath, "git", "git.log")
	if err != nil {
		t.Fatalf("EnableSection() error: %v", err)
	}

	modules := readProfileModules(t, profilePath)
	gitObj := modules["git"].(map[string]interface{})
	disableRaw := gitObj["disable"].([]interface{})
	if len(disableRaw) != 1 || disableRaw[0] != "git.branch" {
		t.Errorf("disable = %v, want [git.branch]", disableRaw)
	}
}

func TestEnableSection_RemovesEmptyDisableKey(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"modules": {"git": {"shell": true, "disable": ["git.log"]}}}`)

	err := EnableSection(profilePath, "git", "git.log")
	if err != nil {
		t.Fatalf("EnableSection() error: %v", err)
	}

	modules := readProfileModules(t, profilePath)
	gitObj := modules["git"].(map[string]interface{})
	if _, exists := gitObj["disable"]; exists {
		t.Error("disable key should be removed when empty")
	}
}

func TestMutateProfile_PreservesUnknownFields(t *testing.T) {
	dir := t.TempDir()
	profilePath := writeTestProfile(t, dir, `{"_comment": "keep me", "custom_field": 42, "modules": {}}`)

	err := Disable(profilePath, "vim")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	raw, _ := os.ReadFile(profilePath)
	var data map[string]interface{}
	json.Unmarshal(raw, &data)

	if data["_comment"] != "keep me" {
		t.Error("_comment field lost")
	}
	if data["custom_field"] != float64(42) {
		t.Error("custom_field lost")
	}
}
