package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"builder/internal/config"
	"builder/internal/deps"
	"builder/internal/nsis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cleanBuild bool

var installerCmd = &cobra.Command{
	Use:   "build-installer",
	Short: "Build full installer",
	Run: func(cmd *cobra.Command, args []string) {
		version := config.GetVersion()
		productName := viper.GetString("product_name")
		companyName := viper.GetString("company_name")
		oldProductName := viper.GetString("old_product_name")
		nsisScript := viper.GetString("nsis_script")
		if nsisScript == "" {
			nsisScript = "installer.nsi"
		}

		fmt.Printf("Building installer for %s version %s\n", productName, version)

		// Create build dir
		buildDir := "build"
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			fmt.Println("Error creating build dir:", err)
			os.Exit(1)
		}

		// Download all dependencies
		packagesDir := filepath.Join(buildDir, "packages")
		pipToolsDir := filepath.Join(buildDir, "pip_wheels")

		if cleanBuild {
			fmt.Println("Cleaning up old packages and tools...")
			os.RemoveAll(packagesDir)
			os.RemoveAll(pipToolsDir)
		}

		reqFile := filepath.Join(config.GetResourcesDir(), config.GetRequirementsFile())

		pkgs, err := deps.ParseReqFile(reqFile)
		if err != nil {
			fmt.Println("Error reading requirements file:", err)
			os.Exit(1)
		}

		var pkgList []string
		for p := range pkgs {
			pkgList = append(pkgList, p)
		}

		fmt.Printf("Downloading %d packages...\n", len(pkgList))
		if err := deps.DownloadDeps(pkgList, packagesDir); err != nil {
			fmt.Println("Error downloading deps:", err)
			os.Exit(1)
		}

		// Also download pip tools (pip, setuptools, wheel)
		fmt.Println("Downloading pip tools...")
		if err := deps.DownloadDeps([]string{"pip", "setuptools", "wheel"}, pipToolsDir); err != nil {
			fmt.Println("Error downloading pip tools:", err)
			os.Exit(1)
		}

		// Compile NSIS
		// Locate script in resources
		resDir := config.GetResourcesDir()
		absResDir, _ := filepath.Abs(resDir)
		scriptPath := filepath.Join(resDir, nsisScript)
		absReqFile, _ := filepath.Abs(reqFile)

		absBuildDir, _ := filepath.Abs(buildDir)
		installerOutput := filepath.Join(absBuildDir, fmt.Sprintf("%sv%s.exe", productName, version))

		defines := map[string]string{
			"PRODUCT_VERSION":  version,
			"PRODUCT_NAME":     productName,
			"OLD_PRODUCT_NAME": oldProductName,
			"COMPANY_NAME":     companyName,
			"INSTALLER_OUTPUT": installerOutput,
			"PACKAGES_DIR":     filepath.Join(absBuildDir, "packages"),
			"PIP_WHEELS_DIR":   filepath.Join(absBuildDir, "pip_wheels"),
			"REQUIREMENTS_FILE": absReqFile,
			"RESOURCES_DIR":    absResDir,
		}

		if err := nsis.CompileNSIS(scriptPath, defines); err != nil {
			fmt.Println("Error compiling NSIS:", err)
			os.Exit(1)
		}

		fmt.Println("Installer build complete. Output:", installerOutput)
	},
}

func init() {
	installerCmd.Flags().BoolVar(&cleanBuild, "clean", true, "Clean up packages directory before downloading")
	rootCmd.AddCommand(installerCmd)
}
