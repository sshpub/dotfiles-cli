package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Module management",
	Long:  "List, enable, disable, install, and manage dotfiles modules.",
}

var moduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all modules with enabled/disabled status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleBrowseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Interactive TUI module browser",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleInfoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "Show module details, sections, and install recipes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleEnableCmd = &cobra.Command{
	Use:   "enable [name]",
	Short: "Enable a module in the profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleDisableCmd = &cobra.Command{
	Use:   "disable [name]",
	Short: "Disable a module in the profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleInstallCmd = &cobra.Command{
	Use:   "install [name]",
	Short: "Install a module's packages",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleOverrideCmd = &cobra.Command{
	Use:   "override [section]",
	Short: "Clone a section to overrides",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleResetCmd = &cobra.Command{
	Use:   "reset [section]",
	Short: "Remove override, restore module default",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Scaffold a new module from template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleValidateCmd = &cobra.Command{
	Use:   "validate [name]",
	Short: "Validate module.json and section guards",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleAddCmd = &cobra.Command{
	Use:   "add [registry/name]",
	Short: "Install a module from a registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update module(s) from registry",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var moduleSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search all registries for modules",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	moduleListCmd.Flags().Bool("interactive", false, "TUI browser with search/toggle")
	moduleInstallCmd.Flags().Bool("all", false, "Install all enabled modules' packages")
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
