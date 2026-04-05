package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update the CLI binary from GitHub releases",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	selfUpdateCmd.Flags().Bool("check", false, "Just check, don't install")
	rootCmd.AddCommand(selfUpdateCmd)
}
