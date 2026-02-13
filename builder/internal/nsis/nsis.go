package nsis

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

func FindMakensis() (string, error) {
	// First check PATH
	path, err := exec.LookPath("makensis")
	if err == nil {
		return path, nil
	}
	
	// Check standard locations on Windows
	if runtime.GOOS == "windows" {
		possiblePaths := []string{
			"C:\\Program Files (x86)\\NSIS\\makensis.exe",
			"C:\\Program Files\\NSIS\\makensis.exe",
			"C:\\NSIS\\makensis.exe",
		}
		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("makensis not found")
}

func GenerateUpgradeScript(fromVer, toVer, outputDir string) (string, error) {
	// Get resources dir for template
	resDir := "resources"
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		resDir = filepath.Dir(configFile)
	}

	tplPath := filepath.Join(resDir, "upgrade_template.nsi")
	content, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", tplPath, err)
	}

	scriptContent := string(content)
	scriptContent = strings.ReplaceAll(scriptContent, "%%FROM_VERSION%%", fromVer)
	scriptContent = strings.ReplaceAll(scriptContent, "%%TO_VERSION%%", toVer)

	destPath := filepath.Join(outputDir, fmt.Sprintf("upgrade_%s_to_%s.nsi", fromVer, toVer))
	err = ioutil.WriteFile(destPath, []byte(scriptContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write script %s: %w", destPath, err)
	}
	return destPath, nil
}

func CompileNSIS(scriptPath string, defines map[string]string) error {
	makensis, err := FindMakensis()
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	// Determine prefix
	prefix := "/"
	if runtime.GOOS == "linux" {
		prefix = "-"
	}

	// Create temp script with BOM and potential injection
	tempScript := strings.TrimSuffix(scriptPath, ".nsi") + ".temp.nsi"
	f, err := os.Create(tempScript)
	if err != nil {
		return err
	}
	
	// Write UTF-8 BOM
	f.Write([]byte{0xEF, 0xBB, 0xBF})
	
	// Inject !addplugindir for Linux
	if runtime.GOOS == "linux" {
		// Use absolute path to resources dir if possible, or relative
		resDir := "resources"
		configFile := viper.ConfigFileUsed()
		if configFile != "" {
			resDir = filepath.Dir(configFile)
		}
		absResDir, _ := filepath.Abs(resDir)
		
		// Escape backslashes for NSIS if on Windows, but this is Linux block
		// NSIS string escaping?
		f.WriteString(fmt.Sprintf("!addplugindir \"%s\"\n", absResDir))
	}
	f.Write(content)
	f.Close()
	defer os.Remove(tempScript)

	// Build arguments
	args := []string{}
	for k, v := range defines {
		args = append(args, fmt.Sprintf("%sD%s=%s", prefix, k, v))
	}
	
	// Add output file flag if specified in defines, wait, NSIS usually takes output file from script `OutFile`
	// But we can override with /X"OutFile ..."
	// Here we rely on script having proper OutFile or using defines.
	// We passed INSTALLER_OUTPUT via defines.

	args = append(args, prefix+"V2")
	args = append(args, tempScript)

	cmd := exec.Command(makensis, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Printf("Compiling: %s %v\n", makensis, args)
	return cmd.Run()
}
