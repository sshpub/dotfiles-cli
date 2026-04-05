package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Modular dotfiles manager",
	Long:  "A modular, cross-platform dotfiles framework CLI.\n\nManages profiles, modules, caches, and symlinks for the sshpub/dotfiles framework.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
