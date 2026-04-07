package profile

import (
	"errors"
	"runtime"
	"testing"
)

func TestDetectPlatform_Smoke(t *testing.T) {
	info := DetectPlatform()

	if info.OS == "" {
		t.Fatal("OS should not be empty")
	}
	if info.Arch == "" {
		t.Fatal("Arch should not be empty")
	}
	if info.Arch != runtime.GOARCH {
		t.Fatalf("Arch = %q, want %q", info.Arch, runtime.GOARCH)
	}

	// On Linux the OS field must be "linux"
	if runtime.GOOS == "linux" && info.OS != "linux" {
		t.Fatalf("OS = %q on linux, want \"linux\"", info.OS)
	}
	if runtime.GOOS == "darwin" && info.OS != "macos" {
		t.Fatalf("OS = %q on darwin, want \"macos\"", info.OS)
	}
}

func TestParseOSRelease(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("parseOSRelease only works on Linux")
	}

	distro, version := parseOSRelease()

	// On any Linux system /etc/os-release should yield a non-empty distro ID.
	if distro == "" {
		t.Fatal("distro should not be empty on Linux")
	}
	t.Logf("distro=%s version=%s", distro, version)
}

func TestCommandExists_Found(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })

	execLookPath = func(name string) (string, error) {
		if name == "fakecmd" {
			return "/usr/bin/fakecmd", nil
		}
		return "", errors.New("not found")
	}

	if !commandExists("fakecmd") {
		t.Fatal("commandExists should return true for fakecmd")
	}
}

func TestCommandExists_NotFound(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })

	execLookPath = func(name string) (string, error) {
		return "", errors.New("not found")
	}

	if commandExists("nonexistent") {
		t.Fatal("commandExists should return false for nonexistent")
	}
}

func TestDetectPackageManager_MacOS_Arm64(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })
	t.Setenv("HOMEBREW_PREFIX", "")

	execLookPath = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", errors.New("not found")
	}

	info := &PlatformInfo{OS: "macos", Arch: "arm64"}
	mgr, prefix := detectPackageManager(info)

	if mgr != "brew" {
		t.Fatalf("manager = %q, want \"brew\"", mgr)
	}
	if prefix != "/opt/homebrew" {
		t.Fatalf("brewPrefix = %q, want \"/opt/homebrew\"", prefix)
	}
}

func TestDetectPackageManager_MacOS_Amd64(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })
	t.Setenv("HOMEBREW_PREFIX", "")

	execLookPath = func(name string) (string, error) {
		if name == "brew" {
			return "/usr/local/bin/brew", nil
		}
		return "", errors.New("not found")
	}

	info := &PlatformInfo{OS: "macos", Arch: "amd64"}
	mgr, prefix := detectPackageManager(info)

	if mgr != "brew" {
		t.Fatalf("manager = %q, want \"brew\"", mgr)
	}
	if prefix != "/usr/local" {
		t.Fatalf("brewPrefix = %q, want \"/usr/local\"", prefix)
	}
}

func TestDetectPackageManager_MacOS_EnvOverride(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })
	t.Setenv("HOMEBREW_PREFIX", "/custom/brew")

	execLookPath = func(name string) (string, error) {
		if name == "brew" {
			return "/custom/brew/bin/brew", nil
		}
		return "", errors.New("not found")
	}

	info := &PlatformInfo{OS: "macos", Arch: "arm64"}
	mgr, prefix := detectPackageManager(info)

	if mgr != "brew" {
		t.Fatalf("manager = %q, want \"brew\"", mgr)
	}
	if prefix != "/custom/brew" {
		t.Fatalf("brewPrefix = %q, want \"/custom/brew\"", prefix)
	}
}

func TestDetectPackageManager_Linux_Apt(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })

	execLookPath = func(name string) (string, error) {
		if name == "apt" {
			return "/usr/bin/apt", nil
		}
		return "", errors.New("not found")
	}

	info := &PlatformInfo{OS: "linux", Arch: "amd64"}
	mgr, prefix := detectPackageManager(info)

	if mgr != "apt" {
		t.Fatalf("manager = %q, want \"apt\"", mgr)
	}
	if prefix != "" {
		t.Fatalf("brewPrefix = %q, want empty", prefix)
	}
}

func TestDetectPackageManager_Linux_DnfPriority(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })

	// Both dnf and yum exist — dnf should win because it comes first.
	execLookPath = func(name string) (string, error) {
		switch name {
		case "dnf":
			return "/usr/bin/dnf", nil
		case "yum":
			return "/usr/bin/yum", nil
		}
		return "", errors.New("not found")
	}

	info := &PlatformInfo{OS: "linux", Arch: "amd64"}
	mgr, _ := detectPackageManager(info)

	if mgr != "dnf" {
		t.Fatalf("manager = %q, want \"dnf\" (should have priority over yum)", mgr)
	}
}

func TestDetectPackageManager_Linux_BrewFallback(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })
	t.Setenv("HOMEBREW_PREFIX", "")

	// No system package manager, only brew.
	execLookPath = func(name string) (string, error) {
		if name == "brew" {
			return "/home/linuxbrew/.linuxbrew/bin/brew", nil
		}
		return "", errors.New("not found")
	}

	info := &PlatformInfo{OS: "linux", Arch: "amd64"}
	mgr, prefix := detectPackageManager(info)

	if mgr != "brew" {
		t.Fatalf("manager = %q, want \"brew\"", mgr)
	}
	if prefix != "/home/linuxbrew/.linuxbrew" {
		t.Fatalf("brewPrefix = %q, want \"/home/linuxbrew/.linuxbrew\"", prefix)
	}
}

func TestDetectPackageManager_NoPkgManager(t *testing.T) {
	orig := execLookPath
	t.Cleanup(func() { execLookPath = orig })

	execLookPath = func(name string) (string, error) {
		return "", errors.New("not found")
	}

	info := &PlatformInfo{OS: "linux", Arch: "amd64"}
	mgr, prefix := detectPackageManager(info)

	if mgr != "" {
		t.Fatalf("manager = %q, want empty", mgr)
	}
	if prefix != "" {
		t.Fatalf("brewPrefix = %q, want empty", prefix)
	}
}
