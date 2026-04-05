package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Profile management",
}

var profileShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current profile",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var profileEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open profile in $EDITOR",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var profileWizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Re-run interactive profile wizard",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

var profileExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export profile to share as example",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}

func init() {
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileEditCmd)
	profileCmd.AddCommand(profileWizardCmd)
	profileCmd.AddCommand(profileExportCmd)
	rootCmd.AddCommand(profileCmd)
}
