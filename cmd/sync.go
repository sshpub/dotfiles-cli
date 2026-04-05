package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync status and background fetch",
}

var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync state",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var syncCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Background fetch and update cache",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	syncCmd.AddCommand(syncStatusCmd)
	syncCmd.AddCommand(syncCheckCmd)
	rootCmd.AddCommand(syncCmd)
}
