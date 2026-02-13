package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"builder/internal/config"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot-version",
	Short: "Snapshot current requirements",
	Run: func(cmd *cobra.Command, args []string) {
		version, _ := cmd.Flags().GetString("version")
		if version == "" {
			version = config.GetVersion()
		}
		if version == "" {
			fmt.Println("No version specified and no current version in config.")
			os.Exit(1)
		}

		resourcesDir := config.GetResourcesDir()
		reqFile := filepath.Join(resourcesDir, config.GetRequirementsFile())
		versionsDir := filepath.Join(resourcesDir, "versions")
		destFile := filepath.Join(versionsDir, fmt.Sprintf("requirements_%s.txt", version))

		if _, err := os.Stat(reqFile); os.IsNotExist(err) {
			fmt.Printf("Requirements file not found: %s
", reqFile)
			os.Exit(1)
		}

		if err := os.MkdirAll(versionsDir, 0755); err != nil {
			fmt.Println("Error creating versions directory:", err)
			os.Exit(1)
		}

		fmt.Printf("Snapshotting %s to %s
", reqFile, destFile)
		if err := copyFile(reqFile, destFile); err != nil {
			fmt.Println("Error copying file:", err)
			os.Exit(1)
		}
		fmt.Println("Snapshot created.")
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
	snapshotCmd.Flags().String("version", "", "Version to snapshot (default: current config version)")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
