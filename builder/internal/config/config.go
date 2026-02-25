package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

func GetVersion() string {
	return viper.GetString("version")
}

func GetRequirementsFile() string {
	// Assume relative to resources/ dir
	return viper.GetString("requirements_file")
}

func GetPyProjectFile() string {
	return viper.GetString("pyproject_file")
}

func GetStaticResources() map[string]string {
	return viper.GetStringMapString("static_resources")
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
