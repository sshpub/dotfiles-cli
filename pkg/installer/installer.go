package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InstallRecipes mirrors module.InstallRecipes for decoupling.
type InstallRecipes struct {
	Brew    []string
	Apt     []string
	Dnf     []string
	Pacman  []string
	Snap    []string
	Zypper  []string
	Inherit string
}

// ResolveRecipes determines the package manager and package list for the
// current platform. It follows the inherit chain if the current platform
// has no direct recipes.
//
// platformOS is "macos", "linux", or "wsl".
// platformPkgManager is "brew", "apt", "dnf", "pacman", "zypper", "yum", or "".
// recipes maps platform name ("macos", "linux", "wsl") to InstallRecipes.
func ResolveRecipes(platformOS, platformPkgManager string, recipes map[string]*InstallRecipes) (manager string, packages []string, err error) {
	if recipes == nil {
		return "", nil, nil
	}

	// Determine which platform key to look up
	platformKey := platformOS
	if platformOS == "wsl" {
		// WSL is treated as linux unless it has its own entry
		if _, ok := recipes["wsl"]; !ok {
			platformKey = "linux"
		}
	}

	r := recipes[platformKey]
	if r == nil {
		return "", nil, nil
	}

	// Follow inherit chain (max depth 3 to prevent loops)
	for i := 0; i < 3 && r.Inherit != ""; i++ {
		inherited := recipes[r.Inherit]
		if inherited == nil {
			break
		}
		r = inherited
	}

	// Match platform package manager to recipe field
	switch platformPkgManager {
	case "brew":
		return "brew", r.Brew, nil
	case "apt":
		return "apt", r.Apt, nil
	case "dnf":
		return "dnf", r.Dnf, nil
	case "pacman":
		return "pacman", r.Pacman, nil
	case "zypper":
		return "zypper", r.Zypper, nil
	case "yum":
		// yum uses dnf recipes as fallback
		if len(r.Dnf) > 0 {
			return "yum", r.Dnf, nil
		}
		return "", nil, nil
	default:
		return "", nil, nil
	}
}

// installCommand is a package-level var for testability.
var installCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// Install runs the package manager to install the given packages.
// If dryRun is true, prints the command instead of executing it.
func Install(manager string, packages []string, dryRun bool) error {
	if len(packages) == 0 {
		return nil
	}

	var args []string
	needsSudo := false

	switch manager {
	case "brew":
		args = append([]string{"install"}, packages...)
	case "apt":
		args = append([]string{"install", "-y"}, packages...)
		needsSudo = true
	case "dnf", "yum":
		args = append([]string{"install", "-y"}, packages...)
		needsSudo = true
	case "pacman":
		args = append([]string{"-S", "--noconfirm"}, packages...)
		needsSudo = true
	case "zypper":
		args = append([]string{"install", "-y"}, packages...)
		needsSudo = true
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	// Prepend sudo if needed and not root
	cmdName := manager
	if needsSudo && os.Getuid() != 0 {
		args = append([]string{cmdName}, args...)
		cmdName = "sudo"
	}

	if dryRun {
		fmt.Printf("  %s %s\n", cmdName, strings.Join(args, " "))
		return nil
	}

	fmt.Printf("Running: %s %s\n", cmdName, strings.Join(args, " "))
	cmd := installCommand(cmdName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
