package installer

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestResolveRecipes_LinuxApt(t *testing.T) {
	recipes := map[string]*InstallRecipes{
		"linux": {Apt: []string{"git", "git-lfs"}},
	}
	mgr, pkgs, err := ResolveRecipes("linux", "apt", recipes)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if mgr != "apt" {
		t.Errorf("manager = %q, want apt", mgr)
	}
	if len(pkgs) != 2 || pkgs[0] != "git" {
		t.Errorf("packages = %v, want [git git-lfs]", pkgs)
	}
}

func TestResolveRecipes_MacosBrew(t *testing.T) {
	recipes := map[string]*InstallRecipes{
		"macos": {Brew: []string{"ripgrep", "fd"}},
	}
	mgr, pkgs, err := ResolveRecipes("macos", "brew", recipes)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if mgr != "brew" || len(pkgs) != 2 {
		t.Errorf("got %q %v, want brew [ripgrep fd]", mgr, pkgs)
	}
}

func TestResolveRecipes_Inherit(t *testing.T) {
	recipes := map[string]*InstallRecipes{
		"linux": {Apt: []string{"git"}},
		"wsl":   {Inherit: "linux"},
	}
	mgr, pkgs, err := ResolveRecipes("wsl", "apt", recipes)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if mgr != "apt" || len(pkgs) != 1 || pkgs[0] != "git" {
		t.Errorf("got %q %v, want apt [git]", mgr, pkgs)
	}
}

func TestResolveRecipes_WSLFallbackToLinux(t *testing.T) {
	recipes := map[string]*InstallRecipes{
		"linux": {Apt: []string{"curl"}},
	}
	mgr, pkgs, err := ResolveRecipes("wsl", "apt", recipes)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if mgr != "apt" || len(pkgs) != 1 {
		t.Errorf("got %q %v, want apt [curl]", mgr, pkgs)
	}
}

func TestResolveRecipes_NilRecipes(t *testing.T) {
	mgr, pkgs, err := ResolveRecipes("linux", "apt", nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if mgr != "" || pkgs != nil {
		t.Errorf("got %q %v, want empty", mgr, pkgs)
	}
}

func TestResolveRecipes_NoPlatform(t *testing.T) {
	recipes := map[string]*InstallRecipes{
		"macos": {Brew: []string{"git"}},
	}
	mgr, pkgs, err := ResolveRecipes("linux", "apt", recipes)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if mgr != "" || pkgs != nil {
		t.Errorf("got %q %v, want empty", mgr, pkgs)
	}
}

func TestResolveRecipes_YumFallbackToDnf(t *testing.T) {
	recipes := map[string]*InstallRecipes{
		"linux": {Dnf: []string{"git"}},
	}
	mgr, pkgs, err := ResolveRecipes("linux", "yum", recipes)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if mgr != "yum" || len(pkgs) != 1 {
		t.Errorf("got %q %v, want yum [git]", mgr, pkgs)
	}
}

func TestInstall_DryRun(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Install("brew", []string{"git", "git-lfs"}, true)

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("error: %v", err)
	}
	output := buf.String()
	if output != "  brew install git git-lfs\n" {
		t.Errorf("output = %q, want \"  brew install git git-lfs\\n\"", output)
	}
}

func TestInstall_EmptyPackages(t *testing.T) {
	err := Install("brew", nil, false)
	if err != nil {
		t.Errorf("Install with empty packages should be no-op, got: %v", err)
	}
}

func TestInstall_UnsupportedManager(t *testing.T) {
	err := Install("nix", []string{"git"}, false)
	if err == nil {
		t.Error("expected error for unsupported manager")
	}
}
