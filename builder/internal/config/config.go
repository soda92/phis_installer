package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Version        string
	ProductName    string `mapstructure:"product_name"`
	RequirementsFile string `mapstructure:"requirements_file"`
	NSISScript     string `mapstructure:"nsis_script"`
}

func GetVersion() string {
	return viper.GetString("version")
}

func GetRequirementsFile() string {
	// Assume relative to resources/ dir
	return viper.GetString("requirements_file")
}

func GetIndexURL() string {
	idx := viper.GetString("index_url")
	if idx == "" {
		return "https://mirrors.tuna.tsinghua.edu.cn/pypi/web/simple"
	}
	return idx
}

func GetResourcesDir() string {
    // If config file is found, assume resources is its dir
    configFile := viper.ConfigFileUsed()
    if configFile != "" {
        return filepath.Dir(configFile)
    }
    return "resources" // Fallback
}
