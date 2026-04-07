package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sshpub/dotfiles-cli/pkg/module"
	"github.com/sshpub/dotfiles-cli/pkg/profile"
	"github.com/spf13/cobra"
)

type checkResult struct {
	status  string // "pass", "warn", "fail"
	message string
	fix     string
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Health check for symlinks, modules, profile, and dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		var results []checkResult

		// Check 1: Profile found
		profilePath, err := profile.FindProfile()
		if err != nil {
			results = append(results, checkResult{"fail", "Error searching for profile", ""})
		} else if profilePath == "" {
			results = append(results, checkResult{"warn", "No profile found — using all-modules default", "dotfiles setup"})
		} else {
			results = append(results, checkResult{"pass", fmt.Sprintf("Profile found: %s", profilePath), ""})
		}

		// Check 2: Profile valid JSON
		var p *profile.Profile
		if profilePath != "" {
			p, err = profile.LoadProfile(profilePath)
			if err != nil {
				results = append(results, checkResult{"fail", fmt.Sprintf("Profile invalid: %s", err), ""})
			} else {
				results = append(results, checkResult{"pass", "Profile valid JSON", ""})
			}
		}

		// Check 3: Dotfiles directory exists
		var dotfilesDir string
		if p != nil {
			dotfilesDir, err = profile.FindDotfilesDir(p)
		} else {
			dotfilesDir = os.Getenv("DOTFILES_DIR")
			if dotfilesDir == "" {
				err = fmt.Errorf("not set")
			}
		}
		if err != nil || dotfilesDir == "" {
			results = append(results, checkResult{"fail", "Dotfiles directory not found", "Set $DOTFILES_DIR or dotfiles_dir in profile"})
		} else if info, statErr := os.Stat(dotfilesDir); statErr != nil || !info.IsDir() {
			results = append(results, checkResult{"fail", fmt.Sprintf("Dotfiles directory missing: %s", dotfilesDir), ""})
		} else {
			results = append(results, checkResult{"pass", fmt.Sprintf("Dotfiles directory: %s", dotfilesDir), ""})
		}

		// Check 4: Modules exist and valid
		if dotfilesDir != "" {
			enabled := []string{}
			if p != nil {
				enabled = p.EnabledModules()
			}

			modules, discoverErr := module.DiscoverModules(dotfilesDir)
			if discoverErr != nil {
				results = append(results, checkResult{"warn", fmt.Sprintf("Cannot read modules: %s", discoverErr), ""})
			} else {
				// Build lookup of discovered modules
				discovered := make(map[string]module.Module)
				for _, m := range modules {
					discovered[m.Name] = m
				}

				if len(enabled) == 0 && p != nil {
					results = append(results, checkResult{"pass", fmt.Sprintf("%d modules discovered (no profile filter)", len(modules)), ""})
				} else if len(enabled) > 0 {
					allExist := true
					for _, name := range enabled {
						if _, ok := discovered[name]; !ok {
							results = append(results, checkResult{"fail", fmt.Sprintf("Module %s: enabled but directory missing", name), ""})
							allExist = false
						}
					}
					if allExist {
						results = append(results, checkResult{"pass", fmt.Sprintf("%d modules enabled, all directories exist", len(enabled)), ""})
					}
				}

				// Check 4b: Module dependencies satisfied
				if len(enabled) > 0 {
					enabledSet := make(map[string]bool)
					for _, name := range enabled {
						enabledSet[name] = true
					}
					for _, name := range enabled {
						m, ok := discovered[name]
						if !ok {
							continue
						}
						for _, dep := range m.Dependencies {
							if !enabledSet[dep] {
								results = append(results, checkResult{
									"fail",
									fmt.Sprintf("Module %s: missing dependency %q", name, dep),
									fmt.Sprintf("dotfiles module enable %s", dep),
								})
							}
						}
					}
				}

				// Check module.json validity
				for _, m := range modules {
					errs := module.ValidateModule(&m)
					if len(errs) > 0 {
						results = append(results, checkResult{
							"warn",
							fmt.Sprintf("Module %s: %s", m.Name, strings.Join(errs, "; ")),
							"",
						})
					}
				}
			}
		}

		// Check 5: Platform cache exists and matches
		dataDir := profile.FindDataDir()
		platformCache := filepath.Join(dataDir, "cache", "platform.sh")
		if _, statErr := os.Stat(platformCache); statErr != nil {
			results = append(results, checkResult{"warn", "Platform cache missing", "dotfiles cache rebuild"})
		} else {
			results = append(results, checkResult{"pass", "Platform cache current", ""})
		}

		// Check 6: Profile cache freshness
		if profilePath != "" {
			profileCache := filepath.Join(dataDir, "cache", "profile.sh")
			profileInfo, profileStatErr := os.Stat(profilePath)
			cacheInfo, cacheStatErr := os.Stat(profileCache)
			if cacheStatErr != nil {
				results = append(results, checkResult{"warn", "Profile cache missing", "dotfiles cache rebuild"})
			} else if profileStatErr == nil && profileInfo.ModTime().After(cacheInfo.ModTime()) {
				results = append(results, checkResult{"warn", "Profile cache stale (profile modified after cache)", "dotfiles cache rebuild"})
			} else if profileStatErr == nil {
				results = append(results, checkResult{"pass", "Profile cache current", ""})
			}
		}

		// Check 7: Symlinks intact
		if dotfilesDir != "" {
			modules, _ := module.DiscoverModules(dotfilesDir)
			brokenLinks := 0
			for _, m := range modules {
				for _, target := range m.Symlinks {
					expanded := os.ExpandEnv(target)
					if _, err := os.Stat(expanded); err != nil {
						brokenLinks++
					}
				}
			}
			if brokenLinks > 0 {
				results = append(results, checkResult{"warn", fmt.Sprintf("%d broken symlinks", brokenLinks), "dotfiles module sync"})
			} else {
				results = append(results, checkResult{"pass", "No broken symlinks", ""})
			}
		}

		// Check 8: Binary architecture
		detected := profile.DetectPlatform()
		binaryArch := runtime.GOARCH
		if binaryArch == detected.Arch {
			results = append(results, checkResult{"pass", fmt.Sprintf("Binary matches platform (%s/%s)", detected.OS, detected.Arch), ""})
		} else {
			results = append(results, checkResult{"warn", fmt.Sprintf("Binary arch %s != detected %s", binaryArch, detected.Arch), ""})
		}

		// Print results
		fmt.Println()
		pass, warn, fail := 0, 0, 0
		for _, r := range results {
			var icon string
			switch r.status {
			case "pass":
				icon = "✓"
				pass++
			case "warn":
				icon = "⚠"
				warn++
			case "fail":
				icon = "✗"
				fail++
			}
			fmt.Printf("  %s %s\n", icon, r.message)
			if r.fix != "" {
				fmt.Printf("    → Fix: %s\n", r.fix)
			}
		}
		fmt.Printf("\n%d passed, %d warnings, %d failures\n", pass, warn, fail)

		if fail > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
