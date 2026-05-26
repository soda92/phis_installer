package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"builder/internal/config"
	"builder/internal/deps"
	"builder/internal/utils"
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

		if err := utils.ValidateVersion(version); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		resourcesDir := config.GetResourcesDir()
		versionsDir := filepath.Join(resourcesDir, "versions")
		destFile := filepath.Join(versionsDir, fmt.Sprintf("requirements_%s.txt", version))

		sourceFile := ""
		pyProject := config.GetPyProjectFile()
		if pyProject != "" {
			sourceFile = pyProject
		} else {
			sourceFile = filepath.Join(resourcesDir, config.GetRequirementsFile())
		}

		if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
			fmt.Printf("Source file not found: %s\n", sourceFile)
			os.Exit(1)
		}

		if err := os.MkdirAll(versionsDir, 0755); err != nil {
			fmt.Println("Error creating versions directory:", err)
			os.Exit(1)
		}

		if pyProject != "" {
			fmt.Printf("Resolving %s to %s\n", sourceFile, destFile)
			// Use versionsDir as temp dir for resolution output
			resolved, err := deps.ResolveReqFile(sourceFile, versionsDir)
			if err != nil {
				fmt.Println("Error resolving dependencies:", err)
				os.Exit(1)
			}
			// Rename resolved file to destFile if it's different
			if resolved != destFile {
				if resolved == sourceFile {
					if err := copyFile(resolved, destFile); err != nil {
						fmt.Println("Error copying file:", err)
						os.Exit(1)
					}
				} else {
					if err := os.Rename(resolved, destFile); err != nil {
						fmt.Println("Error moving resolved file:", err)
						os.Exit(1)
					}
				}
			}
		} else {
			fmt.Printf("Snapshotting %s to %s\n", sourceFile, destFile)
			if err := copyFile(sourceFile, destFile); err != nil {
				fmt.Println("Error copying file:", err)
				os.Exit(1)
			}
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
