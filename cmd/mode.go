package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sshpub/dotfiles-cli/pkg/profile"
	"github.com/spf13/cobra"
)

var modeFlag string

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Mode management (minimal, server, etc.)",
}

var modeShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show mode configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, _, err := loadProfileOrDefault()
		if err != nil {
			return err
		}

		modeName := modeFlag

		if p.Modes == nil || p.Modes[modeName] == nil {
			if modeName == "minimal" {
				fmt.Println("Mode: minimal (hardcoded default)")
				fmt.Println("Triggers: CLAUDE_CODE, CODEX, GEMINI_CLI, OPENCODE, GROK_CLI, CI, GITHUB_ACTIONS, GITLAB_CI")
				fmt.Println("Loads modules: (none — platform + PATH + exports only)")
				return nil
			}
			return fmt.Errorf("mode %q not found in profile", modeName)
		}

		mode := p.Modes[modeName]
		modeType := mode.Type
		if modeType == "" {
			modeType = "include"
		}

		fmt.Printf("Mode: %s (type: %s)\n", modeName, modeType)
		if len(mode.EnvTriggers) > 0 {
			fmt.Printf("Triggers: %s\n", strings.Join(mode.EnvTriggers, ", "))
		}
		if len(mode.IncludeModules) > 0 {
			fmt.Printf("Loads modules: %s\n", strings.Join(mode.IncludeModules, ", "))
		}
		if len(mode.NeverLoad) > 0 {
			fmt.Printf("Never loads: %s\n", strings.Join(mode.NeverLoad, ", "))
		}

		return nil
	},
}

var modeTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Spawn a shell in the named mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		modeName := modeFlag
		fmt.Printf("Spawning shell with DOTFILES_MODE=%s ...\n", modeName)

		c := exec.Command("bash", "-l")
		c.Env = append(os.Environ(), fmt.Sprintf("DOTFILES_MODE=%s", modeName))
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var modeAddTriggerCmd = &cobra.Command{
	Use:   "add-trigger VAR",
	Short: "Add an env trigger to a mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		trigger := args[0]
		return mutateMode(modeFlag, func(mode *profile.Mode) {
			// Don't add duplicates
			for _, t := range mode.EnvTriggers {
				if t == trigger {
					return
				}
			}
			mode.EnvTriggers = append(mode.EnvTriggers, trigger)
		})
	},
}

var modeIncludeCmd = &cobra.Command{
	Use:   "include MODULE",
	Short: "Add module to mode's include list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mod := args[0]
		return mutateMode(modeFlag, func(mode *profile.Mode) {
			for _, m := range mode.IncludeModules {
				if m == mod {
					return
				}
			}
			mode.IncludeModules = append(mode.IncludeModules, mod)
		})
	},
}

var modeExcludeCmd = &cobra.Command{
	Use:   "exclude MODULE",
	Short: "Exclude module from mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mod := args[0]
		return mutateMode(modeFlag, func(mode *profile.Mode) {
			modeType := mode.Type
			if modeType == "" {
				modeType = "include"
			}

			if modeType == "include" {
				// Remove from include list
				filtered := mode.IncludeModules[:0]
				for _, m := range mode.IncludeModules {
					if m != mod {
						filtered = append(filtered, m)
					}
				}
				mode.IncludeModules = filtered
			} else {
				// Add to never_load (no duplicates)
				for _, m := range mode.NeverLoad {
					if m == mod {
						return
					}
				}
				mode.NeverLoad = append(mode.NeverLoad, mod)
			}
		})
	},
}

// loadProfileOrDefault loads the profile and returns it with its path.
// Returns a zero Profile if none found.
func loadProfileOrDefault() (*profile.Profile, string, error) {
	profilePath, err := profile.FindProfile()
	if err != nil {
		return nil, "", err
	}
	if profilePath == "" {
		return &profile.Profile{}, "", nil
	}
	p, err := profile.LoadProfile(profilePath)
	if err != nil {
		return nil, "", err
	}
	return p, profilePath, nil
}

// mutateMode loads profile, applies a mutation to the named mode, saves, and rebuilds cache.
func mutateMode(modeName string, fn func(*profile.Mode)) error {
	profilePath, err := profile.FindProfile()
	if err != nil {
		return err
	}
	if profilePath == "" {
		return fmt.Errorf("no profile found — create one with: dotfiles setup")
	}

	p, err := profile.LoadProfile(profilePath)
	if err != nil {
		return err
	}

	if p.Modes == nil {
		p.Modes = make(map[string]*profile.Mode)
	}
	if p.Modes[modeName] == nil {
		p.Modes[modeName] = &profile.Mode{Type: "include"}
	}

	fn(p.Modes[modeName])

	if err := profile.SaveProfile(profilePath, p); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	fmt.Printf("Updated mode %q in %s\n", modeName, profilePath)

	// Rebuild cache
	dataDir := profile.FindDataDir()
	if err := profile.GenerateProfileCache(dataDir, profilePath, p); err != nil {
		return fmt.Errorf("rebuilding cache: %w", err)
	}
	fmt.Println("Cache rebuilt")

	return nil
}

func init() {
	// Add --mode flag to all subcommands
	for _, cmd := range []*cobra.Command{modeShowCmd, modeTestCmd, modeAddTriggerCmd, modeIncludeCmd, modeExcludeCmd} {
		cmd.Flags().StringVar(&modeFlag, "mode", "minimal", "mode name")
	}
	modeCmd.AddCommand(modeShowCmd)
	modeCmd.AddCommand(modeTestCmd)
	modeCmd.AddCommand(modeAddTriggerCmd)
	modeCmd.AddCommand(modeIncludeCmd)
	modeCmd.AddCommand(modeExcludeCmd)
	rootCmd.AddCommand(modeCmd)
}
