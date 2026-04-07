package cmd

import (
	"fmt"

	"github.com/sshpub/dotfiles-cli/pkg/profile"
	"github.com/spf13/cobra"
)

var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Show detected platform information",
	Run: func(cmd *cobra.Command, args []string) {
		info := profile.DetectPlatform()
		fmt.Println("Platform Information:")
		fmt.Printf("  OS:              %s\n", info.OS)
		fmt.Printf("  Architecture:    %s\n", info.Arch)
		if info.Distro != "" {
			fmt.Printf("  Distribution:    %s %s\n", info.Distro, info.DistroVersion)
		}
		if info.WSL {
			fmt.Println("  WSL:             yes")
		}
		if info.MacOSVersion != "" {
			fmt.Printf("  macOS Version:   %s\n", info.MacOSVersion)
		}
		if info.PkgManager != "" {
			fmt.Printf("  Package Manager: %s\n", info.PkgManager)
		}
		if info.HomebrewPrefix != "" {
			fmt.Printf("  Homebrew Prefix: %s\n", info.HomebrewPrefix)
		}
		if info.Container {
			fmt.Println("  Container:       yes")
		}
	},
}

func init() {
	rootCmd.AddCommand(platformCmd)
}
