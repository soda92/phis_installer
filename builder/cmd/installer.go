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

var cleanBuild bool

var installerCmd = &cobra.Command{
	Use:   "build-installer",
	Short: "Build full installer",
	Run: func(cmd *cobra.Command, args []string) {
		version := config.GetVersion()
		if err := utils.ValidateVersion(version); err != nil {
			fmt.Println("Error invalid version in config:", err)
			os.Exit(1)
		}

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

		sourceFile := ""
		pyProject := config.GetPyProjectFile()
		if pyProject != "" {
			// Resolve relative path to config file if needed? 
			// For now assume relative to CWD or absolute
			sourceFile = pyProject
			fmt.Printf("Using pyproject file: %s\n", sourceFile)
		} else {
			sourceFile = filepath.Join(config.GetResourcesDir(), config.GetRequirementsFile())
			fmt.Printf("Using requirements file: %s\n", sourceFile)
		}

		resolvedReq, err := deps.DownloadReqFile(sourceFile, packagesDir)
		if err != nil {
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
		absResDir, err := filepath.Abs(resDir)
		if err != nil {
			fmt.Println("Error getting absolute path for resources:", err)
			os.Exit(1)
		}

		scriptPath := filepath.Join(resDir, nsisScript)
		absResolvedReq, err := filepath.Abs(resolvedReq)
		if err != nil {
			fmt.Println("Error getting absolute path for requirements:", err)
			os.Exit(1)
		}

		absBuildDir, err := filepath.Abs(buildDir)
		if err != nil {
			fmt.Println("Error getting absolute path for build dir:", err)
			os.Exit(1)
		}
		installerOutput := filepath.Join(absBuildDir, fmt.Sprintf("%sv%s.exe", productName, version))

		defines := map[string]string{
			"PRODUCT_VERSION":  version,
			"PRODUCT_NAME":     productName,
			"OLD_PRODUCT_NAME": oldProductName,
			"COMPANY_NAME":     companyName,
			"INSTALLER_OUTPUT": installerOutput,
			"PACKAGES_DIR":     filepath.Join(absBuildDir, "packages"),
			"PIP_WHEELS_DIR":   filepath.Join(absBuildDir, "pip_wheels"),
			"REQUIREMENTS_FILE": absResolvedReq,
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
