package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"builder/internal/config"
	"builder/internal/deps"
	"builder/internal/nsis"
	"builder/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cleanUpgrade bool

var upgradeCmd = &cobra.Command{
	Use:   "build-upgrade",
	Short: "Build upgrade package",
	Run: func(cmd *cobra.Command, args []string) {
		fromVer, _ := cmd.Flags().GetString("from-ver")
		toVer, _ := cmd.Flags().GetString("to-ver")

		if toVer == "" {
			toVer = config.GetVersion()
		}

		if fromVer == "" {
			fmt.Println("Error: --from-ver is required")
			os.Exit(1)
		}

		if err := utils.ValidateVersion(fromVer); err != nil {
			fmt.Println("Error invalid from-ver:", err)
			os.Exit(1)
		}
		if err := utils.ValidateVersion(toVer); err != nil {
			fmt.Println("Error invalid to-ver:", err)
			os.Exit(1)
		}

		fmt.Printf("Building upgrade from %s to %s\n", fromVer, toVer)

		// 1. Calculate Diff
		diffPkgs, err := deps.GetDiffPackages(fromVer, toVer)
		if err != nil {
			fmt.Println("Error calculating diff:", err)
			os.Exit(1)
		}

		// Create build dir
		buildDir := "build" // Relative to CWD
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			fmt.Println("Error creating build dir:", err)
			os.Exit(1)
		}

		dlDir := filepath.Join(buildDir, fmt.Sprintf("packages_upgrade_%s_to_%s", fromVer, toVer))
		reqFile := filepath.Join(buildDir, fmt.Sprintf("requirements_upgrade_%s_to_%s.txt", fromVer, toVer))

		if cleanUpgrade {
			fmt.Println("Cleaning up old upgrade packages...")
			os.RemoveAll(dlDir)
		}

		// Write requirements file
		f, err := os.Create(reqFile)
		if err != nil {
			fmt.Println("Error creating requirements file:", err)
			os.Exit(1)
		}
		defer f.Close()

		for _, pkg := range diffPkgs {
			if _, err := f.WriteString(pkg + "\n"); err != nil {
				fmt.Println("Error writing to requirements file:", err)
				os.Exit(1)
			}
		}

		if len(diffPkgs) > 0 {
			fmt.Printf("Found %d new packages. Downloading...\n", len(diffPkgs))
			if err := deps.DownloadDeps(diffPkgs, dlDir); err != nil {
				fmt.Println("Error downloading deps:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("No new packages. Creating empty upgrade.")
			os.MkdirAll(dlDir, 0755)
		}

		// 2. Generate NSIS Script
		nsiPath, err := nsis.GenerateUpgradeScript(fromVer, toVer, buildDir)
		if err != nil {
			fmt.Println("Error generating NSIS script:", err)
			os.Exit(1)
		}

		// 3. Compile
		productName := viper.GetString("product_name")
		companyName := viper.GetString("company_name")
		oldProductName := viper.GetString("old_product_name")

		// Use absolute path for output to avoid confusion if cwd changes (though it shouldn't here)
		absBuildDir, err := filepath.Abs(buildDir)
		if err != nil {
			fmt.Println("Error getting absolute path for build dir:", err)
			os.Exit(1)
		}
		installerOutput := filepath.Join(absBuildDir, fmt.Sprintf("%s_升级包_%s_至_%s.exe", productName, fromVer, toVer))

		defines := map[string]string{
			"PRODUCT_NAME":     productName,
			"COMPANY_NAME":     companyName,
			"OLD_PRODUCT_NAME": oldProductName,
			"INSTALLER_OUTPUT": installerOutput,
		}

		if err := nsis.CompileNSIS(nsiPath, defines); err != nil {
			fmt.Println("Error compiling NSIS:", err)
			os.Exit(1)
		}

		fmt.Println("Upgrade build complete. Output:", installerOutput)
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&cleanUpgrade, "clean", true, "Clean up upgrade packages directory before downloading")
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().String("from-ver", "", "Upgrade from version")
	upgradeCmd.Flags().String("to-ver", "", "Upgrade to version (default: current)")
}
