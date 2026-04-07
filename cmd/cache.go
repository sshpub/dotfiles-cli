package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sshpub/dotfiles-cli/pkg/profile"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache management",
}

var cacheRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Regenerate shell cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := profile.FindDataDir()
		cacheDir := filepath.Join(dataDir, "cache")

		// Platform cache
		info := profile.DetectPlatform()
		if err := profile.GeneratePlatformCache(dataDir, info); err != nil {
			return fmt.Errorf("generating platform cache: %w", err)
		}
		fmt.Printf("Written: %s\n", filepath.Join(cacheDir, "platform.sh"))

		// Profile cache
		profilePath, err := profile.FindProfile()
		if err != nil {
			return fmt.Errorf("finding profile: %w", err)
		}

		if profilePath == "" {
			// No profile — clear stale profile cache if it exists
			os.Remove(filepath.Join(cacheDir, "profile.sh"))
			fmt.Println("No profile found — loader will use all-modules default")
			return nil
		}

		p, err := profile.LoadProfile(profilePath)
		if err != nil {
			return fmt.Errorf("loading profile %s: %w", profilePath, err)
		}

		if err := profile.GenerateProfileCache(dataDir, profilePath, p); err != nil {
			return fmt.Errorf("generating profile cache: %w", err)
		}
		fmt.Printf("Written: %s\n", filepath.Join(cacheDir, "profile.sh"))

		return nil
	},
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cached state",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := profile.FindDataDir()
		if err := profile.ClearCache(dataDir); err != nil {
			return fmt.Errorf("clearing cache: %w", err)
		}
		fmt.Printf("Cleared: %s/cache/\n", dataDir)
		return nil
	},
}

func init() {
	cacheCmd.AddCommand(cacheRebuildCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	rootCmd.AddCommand(cacheCmd)
}
