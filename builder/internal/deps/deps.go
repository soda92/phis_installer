package deps

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

// ParseReqFile parses a requirements file into a map for fast lookup
func ParseReqFile(path string) (map[string]struct{}, error) {
	pkgs := make(map[string]struct{})
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pkgs[line] = struct{}{}
	}
	return pkgs, scanner.Err()
}

func GetDiffPackages(fromVer, toVer string) ([]string, error) {
	resDir := viper.GetString("resources_dir") // Ensure this is set in config/viper
	if resDir == "" {
		resDir = "resources" // Fallback
	}
	
	versionsDir := filepath.Join(resDir, "versions")
	fromFile := filepath.Join(versionsDir, fmt.Sprintf("requirements_%s.txt", fromVer))
	toFile := filepath.Join(versionsDir, fmt.Sprintf("requirements_%s.txt", toVer))

	// If toVer is the current version, fallback to main requirements.txt if specific version file doesn't exist
	if toVer == viper.GetString("version") {
		if _, err := os.Stat(toFile); os.IsNotExist(err) {
			toFile = filepath.Join(resDir, viper.GetString("requirements_file"))
		}
	}

	fromPkgs, err := ParseReqFile(fromFile)
	if err != nil {
		// If from file missing, treat as empty set? Or error?
		// Python logic: fell back to tag parsing. Go logic: simpler, just error or empty.
		fmt.Printf("Warning: Could not read from version file %s: %v\n", fromFile, err)
		fromPkgs = make(map[string]struct{})
	}

	toPkgs, err := ParseReqFile(toFile)
	if err != nil {
		return nil, fmt.Errorf("error reading to version file %s: %w", toFile, err)
	}

	var diff []string
	for pkg := range toPkgs {
		if _, exists := fromPkgs[pkg]; !exists {
			diff = append(diff, pkg)
		}
	}
	sort.Strings(diff)
	return diff, nil
}

func DownloadDeps(packages []string, targetDir string) error {
	if len(packages) == 0 {
		return nil
	}
	
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// Create temp requirements file
	tempReq := filepath.Join(targetDir, "temp_reqs.txt")
	f, err := os.Create(tempReq)
	if err != nil {
		return err
	}
	
	for _, pkg := range packages {
		f.WriteString(pkg + "\n")
	}
	f.Close()
	defer os.Remove(tempReq) // Clean up

	indexURL := viper.GetString("index_url")
	if indexURL == "" {
		indexURL = "https://mirrors.tuna.tsinghua.edu.cn/pypi/web/simple"
	}

	// Construct pip command
	// Default to 'python3' on Linux, 'python' on Windows
	pythonExe := "python"
	if runtime.GOOS != "windows" {
		if _, err := exec.LookPath("python3"); err == nil {
			pythonExe = "python3"
		}
	}

	args := []string{"-m", "pip", "download", "-r", tempReq, "-d", targetDir, "-i", indexURL}

	// Add cross-compilation flags if running on Linux
	if runtime.GOOS == "linux" {
		fmt.Println("Detected Linux environment. Adding cross-platform download flags (win_amd64, py3.8).")
		args = append(args, 
			"--platform", "win_amd64", 
			"--only-binary=:all:", 
			"--python-version", "3.8",
			"--no-deps", // Avoid resolving deps for cross-platform stability if just downloading exact packages
		)
	}

	cmd := exec.Command(pythonExe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Printf("Running: %s %v\n", pythonExe, args)
	return cmd.Run()
}
