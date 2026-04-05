package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Show detected platform information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	rootCmd.AddCommand(platformCmd)
}
