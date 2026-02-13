package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "builder",
	Short: "Installer builder tool",
	Long:  `Tool for managing dependencies, versions, and building NSIS installers.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is resources/config.toml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search in resources directory relative to current working directory
		viper.AddConfigPath("resources")
		viper.AddConfigPath("../resources")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv() 

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Config file not found or invalid:", err)
	}
}

// GetResourcesDir returns the absolute path to the resources directory
func GetResourcesDir() string {
    // Assuming we run from project root, resources is just "resources"
    // Or we can derive it from config file location
    configFile := viper.ConfigFileUsed()
    if configFile != "" {
        return filepath.Dir(configFile)
    }
    return "resources"
}
