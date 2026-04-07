package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sshpub/dotfiles-cli/pkg/installer"
	"github.com/sshpub/dotfiles-cli/pkg/module"
	"github.com/sshpub/dotfiles-cli/pkg/profile"
	"github.com/spf13/cobra"
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Module management",
	Long:  "List, enable, disable, install, and manage dotfiles modules.",
}

// --- list ---

var moduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all modules with enabled/disabled status",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, _, dotfilesDir, err := loadContext()
		if err != nil {
			return err
		}

		modules, err := module.DiscoverModules(dotfilesDir)
		if err != nil {
			return err
		}

		enabledSet := make(map[string]bool)
		noProfile := false
		if p != nil {
			for _, name := range p.EnabledModules() {
				enabledSet[name] = true
			}
		} else {
			noProfile = true
		}

		sort.Slice(modules, func(i, j int) bool {
			return modules[i].Name < modules[j].Name
		})

		enabledCount := 0
		for _, m := range modules {
			if noProfile || enabledSet[m.Name] {
				enabledCount++
			}
		}

		fmt.Printf("Modules (%d discovered, %d enabled):\n\n", len(modules), enabledCount)
		fmt.Printf("  %-20s %-10s %s\n", "NAME", "STATUS", "DESCRIPTION")

		for _, m := range modules {
			status := "disabled"
			if noProfile || enabledSet[m.Name] {
				status = "enabled"
			}
			fmt.Printf("  %-20s %-10s %s\n", m.Name, status, m.Description)
		}
		return nil
	},
}

// --- info ---

var moduleInfoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "Show module details, sections, and install recipes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, _, dotfilesDir, err := loadContext()
		if err != nil {
			return err
		}

		mod, err := findModule(dotfilesDir, args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Module: %s (%s)\n", mod.Name, mod.Version)
		fmt.Printf("  %s\n", mod.Description)
		if mod.Author != "" {
			fmt.Printf("  Author: %s\n", mod.Author)
		}

		// Status from profile
		status := "enabled"
		if p != nil {
			enabled := p.EnabledModules()
			found := false
			for _, name := range enabled {
				if name == mod.Name {
					found = true
					break
				}
			}
			if !found {
				status = "disabled"
			}
		}
		fmt.Printf("  Status: %s\n", status)
		fmt.Println()

		// Sections
		if len(mod.Sections) > 0 {
			fmt.Println("Sections:")
			for name, desc := range mod.Sections {
				fmt.Printf("  %-30s %s\n", name, desc)
			}
			fmt.Println()
		}

		// Shell files
		if mod.Shell != nil && len(mod.Shell.LoadOrder) > 0 {
			fmt.Printf("Shell: %s\n", strings.Join(mod.Shell.LoadOrder, ", "))
		}

		// Dependencies
		if len(mod.Dependencies) > 0 {
			fmt.Printf("Dependencies: %s\n", strings.Join(mod.Dependencies, ", "))
		}

		// Platforms
		if len(mod.Platforms) > 0 {
			fmt.Printf("Platforms: %s\n", strings.Join(mod.Platforms, ", "))
		}

		// Install recipes
		if len(mod.Install) > 0 {
			fmt.Println("\nInstall recipes:")
			for platform, recipes := range mod.Install {
				if recipes.Inherit != "" {
					fmt.Printf("  %s: (inherits %s)\n", platform, recipes.Inherit)
				} else {
					parts := []string{}
					if len(recipes.Brew) > 0 {
						parts = append(parts, "brew: "+strings.Join(recipes.Brew, ", "))
					}
					if len(recipes.Apt) > 0 {
						parts = append(parts, "apt: "+strings.Join(recipes.Apt, ", "))
					}
					if len(recipes.Dnf) > 0 {
						parts = append(parts, "dnf: "+strings.Join(recipes.Dnf, ", "))
					}
					if len(recipes.Pacman) > 0 {
						parts = append(parts, "pacman: "+strings.Join(recipes.Pacman, ", "))
					}
					fmt.Printf("  %s: %s\n", platform, strings.Join(parts, " | "))
				}
			}
		}

		// Symlinks
		if len(mod.Symlinks) > 0 {
			fmt.Println("\nSymlinks:")
			for src, dst := range mod.Symlinks {
				fmt.Printf("  %-30s → %s\n", src, dst)
			}
		}

		return nil
	},
}

// --- enable / disable ---

var moduleEnableCmd = &cobra.Command{
	Use:   "enable [name]",
	Short: "Enable a module in the profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, dotfilesDir, err := requireProfile()
		if err != nil {
			return err
		}

		if err := module.Enable(profilePath, args[0], dotfilesDir); err != nil {
			return err
		}

		rebuildCache(profilePath)
		fmt.Printf("Enabled module %q\n", args[0])
		return nil
	},
}

var moduleDisableCmd = &cobra.Command{
	Use:   "disable [name]",
	Short: "Disable a module in the profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, _, err := requireProfile()
		if err != nil {
			return err
		}

		if err := module.Disable(profilePath, args[0]); err != nil {
			return err
		}

		rebuildCache(profilePath)
		fmt.Printf("Disabled module %q\n", args[0])
		return nil
	},
}

// --- install ---

var moduleInstallCmd = &cobra.Command{
	Use:   "install [name]",
	Short: "Install a module's packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, _, dotfilesDir, err := loadContext()
		if err != nil {
			return err
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		all, _ := cmd.Flags().GetBool("all")

		info := profile.DetectPlatform()

		// Determine OS key for recipes
		platformOS := info.OS
		if info.WSL {
			platformOS = "wsl"
		}

		var targets []module.Module
		if all {
			modules, discErr := module.DiscoverModules(dotfilesDir)
			if discErr != nil {
				return discErr
			}
			// Filter to enabled modules
			enabledSet := make(map[string]bool)
			if p != nil {
				for _, name := range p.EnabledModules() {
					enabledSet[name] = true
				}
			}
			for _, m := range modules {
				if p == nil || enabledSet[m.Name] {
					targets = append(targets, m)
				}
			}
		} else {
			if len(args) == 0 {
				return fmt.Errorf("specify a module name or use --all")
			}
			mod, findErr := findModule(dotfilesDir, args[0])
			if findErr != nil {
				return findErr
			}
			targets = []module.Module{*mod}
		}

		for _, mod := range targets {
			recipes := convertRecipes(mod.Install)
			mgr, pkgs, resolveErr := installer.ResolveRecipes(platformOS, info.PkgManager, recipes)
			if resolveErr != nil {
				return resolveErr
			}
			if mgr == "" || len(pkgs) == 0 {
				if !all {
					fmt.Printf("No install recipes for %s on %s/%s\n", mod.Name, platformOS, info.PkgManager)
				}
				continue
			}

			fmt.Printf("\n%s:\n", mod.Name)
			if err := installer.Install(mgr, pkgs, dryRun); err != nil {
				return fmt.Errorf("installing %s: %w", mod.Name, err)
			}
		}

		return nil
	},
}

// --- validate ---

var moduleValidateCmd = &cobra.Command{
	Use:   "validate [name]",
	Short: "Validate module.json and section guards",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, dotfilesDir, err := loadContext()
		if err != nil {
			return err
		}

		var targets []module.Module
		if len(args) > 0 {
			mod, findErr := findModule(dotfilesDir, args[0])
			if findErr != nil {
				return findErr
			}
			targets = []module.Module{*mod}
		} else {
			modules, discErr := module.DiscoverModules(dotfilesDir)
			if discErr != nil {
				return discErr
			}
			targets = modules
		}

		hasErrors := false
		for _, mod := range targets {
			schemaErrs := module.ValidateModule(&mod)
			guardWarns := module.ValidateSectionGuards(&mod)
			allIssues := append(schemaErrs, guardWarns...)

			if len(allIssues) == 0 {
				fmt.Printf("  ✓ %s\n", mod.Name)
			} else {
				hasErrors = true
				fmt.Printf("  ✗ %s\n", mod.Name)
				for _, issue := range allIssues {
					fmt.Printf("    - %s\n", issue)
				}
			}
		}

		if hasErrors {
			os.Exit(1)
		}
		return nil
	},
}

// --- create ---

var moduleCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Scaffold a new module from template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, dotfilesDir, err := loadContext()
		if err != nil {
			return err
		}

		modDir, err := module.Scaffold(dotfilesDir, args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Created module at %s\n", modDir)
		fmt.Println("Files:")
		fmt.Println("  module.json   — edit name, description, sections, install recipes")
		fmt.Println("  aliases.sh    — add your shell aliases and functions")
		fmt.Println("  CLAUDE.md     — describe the module for AI assistants")
		return nil
	},
}

// --- override ---

var moduleOverrideCmd = &cobra.Command{
	Use:   "override [section]",
	Short: "Clone a section to overrides",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, dotfilesDir, err := requireProfile()
		if err != nil {
			return err
		}

		section := args[0]
		local, _ := cmd.Flags().GetBool("local")
		disable, _ := cmd.Flags().GetBool("disable")

		// Parse module name from section (e.g. "git.shortcuts" → "git")
		moduleName := strings.SplitN(section, ".", 2)[0]

		if disable {
			if err := module.DisableSection(profilePath, moduleName, section); err != nil {
				return err
			}
			rebuildCache(profilePath)
			fmt.Printf("Disabled section %q\n", section)
			return nil
		}

		// Find module and extract section
		mod, findErr := findModule(dotfilesDir, moduleName)
		if findErr != nil {
			return findErr
		}

		code, sourceFile, extractErr := module.ExtractSection(mod, section)
		if extractErr != nil {
			return extractErr
		}

		// Determine override directory
		overrideDir := filepath.Join(dotfilesDir, "overrides")
		if local {
			overrideDir = filepath.Join(profile.FindDataDir(), "local")
		}

		if err := module.WriteOverride(overrideDir, section, code, moduleName, sourceFile); err != nil {
			return err
		}

		fmt.Printf("Override written: %s/%s.sh\n", overrideDir, section)
		fmt.Println("Edit this file to customize. Delete to restore module default.")
		return nil
	},
}

// --- reset ---

var moduleResetCmd = &cobra.Command{
	Use:   "reset [section]",
	Short: "Remove override, restore module default",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, dotfilesDir, err := requireProfile()
		if err != nil {
			return err
		}

		section := args[0]
		moduleName := strings.SplitN(section, ".", 2)[0]

		repoDir := filepath.Join(dotfilesDir, "overrides")
		localDir := filepath.Join(profile.FindDataDir(), "local")

		removed := module.ResetOverride(repoDir, localDir, section)

		// Also remove from disable list
		if err := module.EnableSection(profilePath, moduleName, section); err != nil {
			return err
		}

		rebuildCache(profilePath)

		if len(removed) > 0 {
			for _, path := range removed {
				fmt.Printf("Removed: %s\n", path)
			}
		}
		fmt.Printf("Section %q restored to module default\n", section)
		return nil
	},
}

// --- stubs for registry commands (deferred to #30) ---

var moduleBrowseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Interactive TUI module browser",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not implemented — see issue #30 (registry)")
	},
}

var moduleAddCmd = &cobra.Command{
	Use:   "add [registry/name]",
	Short: "Install a module from a registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not implemented — see issue #30 (registry)")
	},
}

var moduleUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update module(s) from registry",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not implemented — see issue #30 (registry)")
	},
}

var moduleSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search all registries for modules",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not implemented — see issue #30 (registry)")
	},
}

// --- helpers ---

// loadContext loads profile and resolves dotfiles dir. Returns nil profile if none found.
func loadContext() (*profile.Profile, string, string, error) {
	profilePath, err := profile.FindProfile()
	if err != nil {
		return nil, "", "", err
	}

	var p *profile.Profile
	if profilePath != "" {
		p, err = profile.LoadProfile(profilePath)
		if err != nil {
			return nil, "", "", fmt.Errorf("loading profile: %w", err)
		}
	}

	dotfilesDir, err := profile.FindDotfilesDir(p)
	if err != nil {
		return nil, "", "", err
	}

	return p, profilePath, dotfilesDir, nil
}

// requireProfile loads profile and dotfiles dir, erroring if no profile found.
func requireProfile() (profilePath string, dotfilesDir string, err error) {
	p, pp, dd, err := loadContext()
	if err != nil {
		return "", "", err
	}
	if pp == "" {
		return "", "", fmt.Errorf("no profile found — create one with: dotfiles setup")
	}
	_ = p
	return pp, dd, nil
}

// findModule discovers all modules and returns the one matching name.
func findModule(dotfilesDir string, name string) (*module.Module, error) {
	modules, err := module.DiscoverModules(dotfilesDir)
	if err != nil {
		return nil, err
	}
	for _, m := range modules {
		if m.Name == name {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("module %q not found in %s/modules/", name, dotfilesDir)
}

// rebuildCache rebuilds the profile cache silently.
func rebuildCache(profilePath string) {
	p, err := profile.LoadProfile(profilePath)
	if err != nil {
		return
	}
	dataDir := profile.FindDataDir()
	profile.GenerateProfileCache(dataDir, profilePath, p)
}

// convertRecipes converts module.InstallRecipes to installer.InstallRecipes.
func convertRecipes(modRecipes map[string]*module.InstallRecipes) map[string]*installer.InstallRecipes {
	if modRecipes == nil {
		return nil
	}
	result := make(map[string]*installer.InstallRecipes, len(modRecipes))
	for platform, r := range modRecipes {
		result[platform] = &installer.InstallRecipes{
			Brew:    r.Brew,
			Apt:     r.Apt,
			Dnf:     r.Dnf,
			Pacman:  r.Pacman,
			Snap:    r.Snap,
			Zypper:  r.Zypper,
			Inherit: r.Inherit,
		}
	}
	return result
}

func init() {
	moduleInstallCmd.Flags().Bool("all", false, "Install all enabled modules' packages")
	moduleInstallCmd.Flags().Bool("dry-run", false, "Print commands instead of executing")
	moduleOverrideCmd.Flags().Bool("local", false, "Clone to ~/.dotfiles/local/ instead")
	moduleOverrideCmd.Flags().Bool("disable", false, "Just disable, don't clone")
	moduleUpdateCmd.Flags().Bool("all", false, "Update all registry-installed modules")
	moduleUpdateCmd.Flags().Bool("check", false, "Show available updates without applying")

	moduleCmd.AddCommand(moduleListCmd)
	moduleCmd.AddCommand(moduleBrowseCmd)
	moduleCmd.AddCommand(moduleInfoCmd)
	moduleCmd.AddCommand(moduleEnableCmd)
	moduleCmd.AddCommand(moduleDisableCmd)
	moduleCmd.AddCommand(moduleInstallCmd)
	moduleCmd.AddCommand(moduleOverrideCmd)
	moduleCmd.AddCommand(moduleResetCmd)
	moduleCmd.AddCommand(moduleCreateCmd)
	moduleCmd.AddCommand(moduleValidateCmd)
	moduleCmd.AddCommand(moduleAddCmd)
	moduleCmd.AddCommand(moduleUpdateCmd)
	moduleCmd.AddCommand(moduleSearchCmd)
	rootCmd.AddCommand(moduleCmd)
}
