package module

import (
	"os"
	"path/filepath"
	"testing"
)

func writeModuleJSON(t *testing.T, dir, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "module.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const validJSON = `{
  "name": "git",
  "version": "1.0.0",
  "description": "Git configuration and aliases",
  "platforms": ["macos", "linux"]
}`

func TestLoadModule_Valid(t *testing.T) {
	dir := t.TempDir()
	path := writeModuleJSON(t, dir, validJSON)

	mod, err := LoadModule(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod.Name != "git" {
		t.Errorf("expected name %q, got %q", "git", mod.Name)
	}
	if mod.Version != "1.0.0" {
		t.Errorf("expected version %q, got %q", "1.0.0", mod.Version)
	}
	if mod.Description != "Git configuration and aliases" {
		t.Errorf("unexpected description: %s", mod.Description)
	}
	if len(mod.Platforms) != 2 {
		t.Errorf("expected 2 platforms, got %d", len(mod.Platforms))
	}
}

func TestLoadModule_MissingFields(t *testing.T) {
	dir := t.TempDir()
	path := writeModuleJSON(t, dir, `{"name": "bare"}`)

	mod, err := LoadModule(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Missing fields should be zero values, not errors from LoadModule.
	if mod.Version != "" {
		t.Errorf("expected empty version, got %q", mod.Version)
	}
	if mod.Description != "" {
		t.Errorf("expected empty description, got %q", mod.Description)
	}
}

func TestLoadModule_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := writeModuleJSON(t, dir, `{invalid json}`)

	_, err := LoadModule(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestValidateModule_Valid(t *testing.T) {
	mod := &Module{
		Name:        "git",
		Version:     "1.0.0",
		Description: "Git configuration and aliases",
		Platforms:   []string{"macos", "linux"},
	}
	errs := ValidateModule(mod)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateModule_MissingRequired(t *testing.T) {
	mod := &Module{}
	errs := ValidateModule(mod)

	expected := map[string]bool{
		"name is required":        false,
		"version is required":     false,
		"description is required": false,
	}
	for _, e := range errs {
		if _, ok := expected[e]; ok {
			expected[e] = true
		}
	}
	for msg, found := range expected {
		if !found {
			t.Errorf("expected error %q not found in %v", msg, errs)
		}
	}
}

func TestValidateModule_BadName(t *testing.T) {
	mod := &Module{
		Name:        "Git-Config",
		Version:     "1.0.0",
		Description: "desc",
	}
	errs := ValidateModule(mod)
	if len(errs) == 0 {
		t.Fatal("expected validation error for uppercase name")
	}
	found := false
	for _, e := range errs {
		if e == `name "Git-Config" must match ^[a-z][a-z0-9-]*$` {
			found = true
		}
	}
	if !found {
		t.Errorf("expected name pattern error, got %v", errs)
	}
}

func TestValidateModule_BadPlatform(t *testing.T) {
	mod := &Module{
		Name:        "test",
		Version:     "1.0.0",
		Description: "desc",
		Platforms:   []string{"windows"},
	}
	errs := ValidateModule(mod)
	if len(errs) == 0 {
		t.Fatal("expected validation error for bad platform")
	}
	found := false
	for _, e := range errs {
		if e == `platform "windows" must be macos, linux, or wsl` {
			found = true
		}
	}
	if !found {
		t.Errorf("expected platform error, got %v", errs)
	}
}

func TestValidateModule_BadDependency(t *testing.T) {
	mod := &Module{
		Name:         "test",
		Version:      "1.0.0",
		Description:  "desc",
		Dependencies: []string{"Valid_Dep"},
	}
	errs := ValidateModule(mod)
	if len(errs) == 0 {
		t.Fatal("expected validation error for bad dependency name")
	}
	found := false
	for _, e := range errs {
		if e == `dependency "Valid_Dep" is not a valid module name` {
			found = true
		}
	}
	if !found {
		t.Errorf("expected dependency error, got %v", errs)
	}
}

func TestDiscoverModules(t *testing.T) {
	root := t.TempDir()
	modulesDir := filepath.Join(root, "modules")

	// Module 1: git
	writeModuleJSON(t, filepath.Join(modulesDir, "git"), `{
		"name": "git",
		"version": "1.0.0",
		"description": "Git config"
	}`)

	// Module 2: ssh
	writeModuleJSON(t, filepath.Join(modulesDir, "ssh"), `{
		"name": "ssh",
		"version": "0.1.0",
		"description": "SSH config"
	}`)

	// Directory without module.json — should be skipped
	if err := os.MkdirAll(filepath.Join(modulesDir, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}

	modules, err := DiscoverModules(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(modules))
	}

	names := map[string]bool{}
	for _, m := range modules {
		names[m.Name] = true
		if m.Dir == "" {
			t.Errorf("module %q has empty Dir", m.Name)
		}
	}
	if !names["git"] {
		t.Error("expected git module to be discovered")
	}
	if !names["ssh"] {
		t.Error("expected ssh module to be discovered")
	}
}
