package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"builder/internal/config"
	"github.com/spf13/cobra"
)

var downloadResourcesCmd = &cobra.Command{
	Use:   "download-resources",
	Short: "Download static resources (Python, VC Redist)",
	Run: func(cmd *cobra.Command, args []string) {
		resDir := config.GetResourcesDir()
		
		resources := map[string]string{
			"python-3.8.10-embed-amd64.zip": "https://www.python.org/ftp/python/3.8.10/python-3.8.10-embed-amd64.zip",
			"VC_redist2015-2022.x64.exe":    "https://aka.ms/vs/17/release/vc_redist.x64.exe",
		}

		for filename, url := range resources {
			destPath := filepath.Join(resDir, filename)
			if _, err := os.Stat(destPath); err == nil {
				fmt.Printf("Resource %s already exists. Skipping.\n", filename)
				continue
			}

			fmt.Printf("Downloading %s from %s...\n", filename, url)
			if err := downloadFile(destPath, url); err != nil {
				fmt.Printf("Error downloading %s: %v\n", filename, err)
				os.Exit(1)
			}
			fmt.Println("Download complete.")
		}
	},
}

func init() {
	rootCmd.AddCommand(downloadResourcesCmd)
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
