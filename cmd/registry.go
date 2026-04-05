package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Registry management",
}

var registryAddCmd = &cobra.Command{
	Use:   "add [name] [url]",
	Short: "Add a module registry",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured registries",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var registryRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var registrySyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Pull latest from all registries",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	registryAddCmd.Flags().Bool("private", false, "Mark as private (SSH auth)")
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registrySyncCmd)
	rootCmd.AddCommand(registryCmd)
}
