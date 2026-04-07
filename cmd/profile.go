package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sshpub/dotfiles-cli/pkg/profile"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Profile management",
}

var profileShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, err := profile.FindProfile()
		if err != nil {
			return err
		}
		if profilePath == "" {
			fmt.Println("No profile found — using all-modules default")
			return nil
		}

		p, err := profile.LoadProfile(profilePath)
		if err != nil {
			return fmt.Errorf("loading profile: %w", err)
		}

		fmt.Printf("Profile: %s\n", profilePath)

		dotfilesDir, err := profile.FindDotfilesDir(p)
		if err == nil {
			fmt.Printf("Dotfiles: %s\n", dotfilesDir)
		}

		fmt.Println()

		// Role
		if len(p.Role) > 0 {
			fmt.Printf("Role: %s\n", strings.Join(p.Role, ", "))
		}

		// Modules
		enabled := p.EnabledModules()
		if len(enabled) > 0 {
			fmt.Printf("Modules (%d enabled):\n", len(enabled))
			fmt.Printf("  %s\n", strings.Join(enabled, ", "))
		}

		disabled := p.DisabledSections()
		if len(disabled) > 0 {
			fmt.Printf("Disabled sections: %s\n", strings.Join(disabled, ", "))
		}

		// Modes
		if len(p.Modes) > 0 {
			fmt.Println("Modes:")
			for name, mode := range p.Modes {
				modeType := mode.Type
				if modeType == "" {
					modeType = "include"
				}
				triggers := strings.Join(mode.EnvTriggers, ", ")
				fmt.Printf("  %s (%s)", name, modeType)
				if triggers != "" {
					fmt.Printf(": triggers %s", triggers)
				}
				if len(mode.IncludeModules) > 0 {
					fmt.Printf(" → loads %s", strings.Join(mode.IncludeModules, ", "))
				}
				if len(mode.NeverLoad) > 0 {
					fmt.Printf(" → skips %s", strings.Join(mode.NeverLoad, ", "))
				}
				fmt.Println()
			}
		}

		// Git
		if p.Git != nil && (p.Git.Name != "" || p.Git.Email != "") {
			fmt.Printf("Git: %s <%s>\n", p.Git.Name, p.Git.Email)
		}

		return nil
	},
}

var profileEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open profile in $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, err := profile.FindProfile()
		if err != nil {
			return err
		}
		if profilePath == "" {
			return fmt.Errorf("no profile found — create one with: dotfiles setup")
		}

		editor := findEditor()
		if editor == "" {
			return fmt.Errorf("no editor found: set $EDITOR or $VISUAL")
		}

		// Open editor
		c := exec.Command(editor, profilePath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %w", err)
		}

		// Auto-rebuild cache after edit
		fmt.Println("Rebuilding cache...")
		return cacheRebuildCmd.RunE(cacheRebuildCmd, nil)
	},
}

var profileWizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Interactive profile wizard",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not implemented — use `dotfiles setup`")
	},
}

var profileExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export profile (redacted) to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, err := profile.FindProfile()
		if err != nil {
			return err
		}
		if profilePath == "" {
			return fmt.Errorf("no profile found")
		}

		p, err := profile.LoadProfile(profilePath)
		if err != nil {
			return err
		}

		// Redact git identity
		if p.Git != nil {
			if p.Git.Name != "" {
				p.Git.Name = "Your Name"
			}
			if p.Git.Email != "" {
				p.Git.Email = "you@example.com"
			}
		}

		data, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

// findEditor returns the first available editor from the fallback chain.
func findEditor() string {
	for _, env := range []string{"EDITOR", "VISUAL"} {
		if e := os.Getenv(env); e != "" {
			return e
		}
	}
	for _, name := range []string{"vim", "vi", "nano"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

func init() {
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileEditCmd)
	profileCmd.AddCommand(profileWizardCmd)
	profileCmd.AddCommand(profileExportCmd)
	rootCmd.AddCommand(profileCmd)
}
