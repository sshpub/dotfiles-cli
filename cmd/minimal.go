package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var minimalCmd = &cobra.Command{
	Use:   "minimal",
	Short: "Minimal mode management",
}

var minimalShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show what loads in minimal mode",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var minimalTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Spawn a minimal shell to try it",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var minimalAddTriggerCmd = &cobra.Command{
	Use:   "add-trigger [ENV_VAR]",
	Short: "Add a new AI tool trigger",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var minimalIncludeCmd = &cobra.Command{
	Use:   "include [module]",
	Short: "Add module to minimal mode",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var minimalExcludeCmd = &cobra.Command{
	Use:   "exclude [module]",
	Short: "Remove module from minimal mode",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	minimalCmd.AddCommand(minimalShowCmd)
	minimalCmd.AddCommand(minimalTestCmd)
	minimalCmd.AddCommand(minimalAddTriggerCmd)
	minimalCmd.AddCommand(minimalIncludeCmd)
	minimalCmd.AddCommand(minimalExcludeCmd)
	rootCmd.AddCommand(minimalCmd)
}
