package deps

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"builder/internal/config"
	"github.com/spf13/viper"
)

// ParseReqFile returns a set of package names (lowercased)
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
		// Basic parsing
		pkgs[line] = struct{}{}
	}
	return pkgs, scanner.Err()
}

// GetDiffPackages returns packages in toVer but not fromVer
func GetDiffPackages(fromVer, toVer string) ([]string, error) {
	resDir := config.GetResourcesDir()
	versionsDir := filepath.Join(resDir, "versions")

	fromFile := filepath.Join(versionsDir, fmt.Sprintf("requirements_%s.txt", fromVer))
	toFile := filepath.Join(versionsDir, fmt.Sprintf("requirements_%s.txt", toVer))

	// Fallback for toVer if it's current
	if toVer == config.GetVersion() {
		if _, err := os.Stat(toFile); os.IsNotExist(err) {
			toFile = filepath.Join(resDir, config.GetRequirementsFile())
		}
	}

	fromPkgs, err := ParseReqFile(fromFile)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", fromFile, err)
	}

	toPkgs, err := ParseReqFile(toFile)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", toFile, err)
	}

	diff := []string{}
	for pkg := range toPkgs {
		if _, exists := fromPkgs[pkg]; !exists {
			diff = append(diff, pkg)
		}
	}
	// Sort for deterministic output? Go maps are unordered.
	return diff, nil
}

// DownloadDeps downloads wheels using pip
func DownloadDeps(packages []string, targetDir string) error {
	if len(packages) == 0 {
		return nil
	}
	
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	tempReq := filepath.Join(targetDir, "temp_reqs.txt")
	f, err := os.Create(tempReq)
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		f.WriteString(pkg + "
")
	}
	f.Close()
	defer os.Remove(tempReq)

	indexURL := config.GetIndexURL()
	
	// pip command
	args := []string{"-m", "pip", "download", "-r", tempReq, "-d", targetDir, "-i", indexURL}

	// Cross-compile logic
	if runtime.GOOS == "linux" {
		fmt.Println("Detected Linux environment. Adding flags for Windows (win_amd64, python 3.8) cross-download.")
		args = append(args, "--platform", "win_amd64", "--only-binary=:all:", "--python-version", "3.8", "--no-deps")
	}

	cmd := exec.Command("python3", args...) 
    // Fallback if python3 not found? usually python on Windows.
    if _, err := exec.LookPath("python3"); err != nil {
         cmd = exec.Command("python", args...)
    }
    
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
