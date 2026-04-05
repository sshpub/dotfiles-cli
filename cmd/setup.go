package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "First-run setup wizard",
	Long:  "Interactive setup wizard for configuring dotfiles on a new machine.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	setupCmd.Flags().Bool("non-interactive", false, "Run setup without prompts (for automation)")
	rootCmd.AddCommand(setupCmd)
}
