package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull latest, rebuild cache, re-link",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	updateCmd.Flags().Bool("check", false, "Just check, don't apply")
	updateCmd.Flags().Bool("diff", false, "Show what changed before applying")
	updateCmd.Flags().Bool("core", false, "Fetch upstream dotfiles-core updates")
	rootCmd.AddCommand(updateCmd)
}
