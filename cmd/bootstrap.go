package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Create symlinks for core and enabled modules",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	bootstrapCmd.Flags().Bool("force", false, "Skip confirmation prompts")
	bootstrapCmd.Flags().Bool("dry-run", false, "Show what would be linked")
	rootCmd.AddCommand(bootstrapCmd)
}
