package profile

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PlatformInfo holds detected platform details.
// Matches the variables exported by core/platform.sh.
type PlatformInfo struct {
	OS             string // "macos", "linux", "windows"
	Arch           string // runtime.GOARCH value
	WSL            bool
	Distro         string // e.g. "ubuntu"
	DistroVersion  string // e.g. "25.10"
	PkgManager     string // "apt", "dnf", "brew", etc.
	Container      bool
	HomebrewPrefix string
	MacOSVersion   string
}

var execCommand = func(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}

var execLookPath = exec.LookPath

// DetectPlatform detects the current platform, matching core/platform.sh logic.
func DetectPlatform() *PlatformInfo {
	info := &PlatformInfo{
		Arch: runtime.GOARCH,
	}

	// OS
	switch runtime.GOOS {
	case "darwin":
		info.OS = "macos"
	case "linux":
		info.OS = "linux"
	case "windows":
		info.OS = "windows"
	default:
		info.OS = "unknown"
	}

	// WSL detection
	if info.OS == "linux" {
		if data, err := os.ReadFile("/proc/version"); err == nil {
			if strings.Contains(strings.ToLower(string(data)), "microsoft") {
				info.WSL = true
			}
		}
	}

	// Distro (Linux)
	if info.OS == "linux" {
		info.Distro, info.DistroVersion = parseOSRelease()
	}

	// macOS version
	if info.OS == "macos" {
		info.MacOSVersion = detectMacOSVersion()
	}

	// Container detection
	if _, err := os.Stat("/.dockerenv"); err == nil {
		info.Container = true
	} else if _, err := os.Stat("/run/.containerenv"); err == nil {
		info.Container = true
	}

	// Package manager
	info.PkgManager, info.HomebrewPrefix = detectPackageManager(info)

	return info
}

// parseOSRelease reads /etc/os-release for distro ID and version.
func parseOSRelease() (distro, version string) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		val = strings.Trim(val, "\"")
		switch key {
		case "ID":
			distro = val
		case "VERSION_ID":
			version = val
		}
	}
	return distro, version
}

// detectMacOSVersion runs sw_vers -productVersion.
func detectMacOSVersion() string {
	out, err := execCommand("sw_vers", "-productVersion")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// detectPackageManager finds the system package manager.
func detectPackageManager(info *PlatformInfo) (manager, brewPrefix string) {
	if info.OS == "macos" {
		if commandExists("brew") {
			manager = "brew"
			if p := os.Getenv("HOMEBREW_PREFIX"); p != "" {
				brewPrefix = p
			} else if info.Arch == "arm64" {
				brewPrefix = "/opt/homebrew"
			} else {
				brewPrefix = "/usr/local"
			}
		}
		return
	}

	if info.OS == "linux" {
		for _, candidate := range []string{"apt", "dnf", "pacman", "zypper", "yum"} {
			if commandExists(candidate) {
				return candidate, ""
			}
		}
		if commandExists("brew") {
			if p := os.Getenv("HOMEBREW_PREFIX"); p != "" {
				brewPrefix = p
			} else {
				brewPrefix = "/home/linuxbrew/.linuxbrew"
			}
			return "brew", brewPrefix
		}
	}

	return "", ""
}

// commandExists checks if a command is in PATH.
func commandExists(name string) bool {
	_, err := execLookPath(name)
	return err == nil
}
