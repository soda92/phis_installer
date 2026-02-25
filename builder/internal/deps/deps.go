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

	"builder/internal/config"
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
	resDir := config.GetResourcesDir()
	
	versionsDir := filepath.Join(resDir, "versions")
	fromFile := filepath.Join(versionsDir, fmt.Sprintf("requirements_%s.txt", fromVer))
	toFile := filepath.Join(versionsDir, fmt.Sprintf("requirements_%s.txt", toVer))

	if toVer == config.GetVersion() {
		if _, err := os.Stat(toFile); os.IsNotExist(err) {
			pyProject := config.GetPyProjectFile()
			if pyProject != "" {
				// Resolve it to a temp file in build dir
				fmt.Printf("Resolving current pyproject file for diff: %s\n", pyProject)
				tempDir := "build/temp_diff"
				resolved, err := ResolveReqFile(pyProject, tempDir)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve pyproject for diff: %w", err)
				}
				toFile = resolved
			} else {
				toFile = filepath.Join(resDir, config.GetRequirementsFile())
			}
		}
	}

	fromPkgs, err := ParseReqFile(fromFile)
	if err != nil {
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

	// 1. Create input requirements file
	reqIn := filepath.Join(targetDir, "requirements.in")
	f, err := os.Create(reqIn)
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		f.WriteString(pkg + "\n")
	}
	f.Close()
	
	_, err = DownloadReqFile(reqIn, targetDir)
	return err
}

func DownloadReqFile(reqFile, targetDir string) (string, error) {
	resolvedReq, err := ResolveReqFile(reqFile, targetDir)
	if err != nil {
		return "", err
	}

	indexURL := config.GetIndexURL()

	// 3. Download using pip
	pythonExe := "python"
	if runtime.GOOS != "windows" {
		if _, err := exec.LookPath("python3"); err == nil {
			pythonExe = "python3"
		}
	}

	args := []string{"-m", "pip", "download", "-r", resolvedReq, "-d", targetDir, "-i", indexURL}

	if runtime.GOOS == "linux" {
		// Use --no-deps because we hopefully resolved everything or are forced to
		args = append(args, 
			"--platform", "win_amd64", 
			"--python-version", "3.8",
			"--no-deps", 
		)
	}

	cmd := exec.Command(pythonExe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Printf("Downloading: %s %v\n", pythonExe, args)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return resolvedReq, nil
}

func ResolveReqFile(reqFile, targetDir string) (string, error) {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	indexURL := config.GetIndexURL()

	// Resolve dependencies using 'uv' if available (preferred for cross-platform)
	// or fallback to pip (risky/limited).
	resolvedReq := reqFile // Default to using input as is
	
	uvPath, err := exec.LookPath("uv")
	if err == nil && runtime.GOOS == "linux" {
		fmt.Println("Resolving dependencies with uv (cross-platform target: win_amd64, python 3.8)...")
		resolvedReq = filepath.Join(targetDir, "requirements.txt")
		
		uvArgs := []string{"pip", "compile", 
			reqFile, 
			"-o", resolvedReq, 
			"--python-version", "3.8", 
			"--python-platform", "x86_64-pc-windows-msvc",
			"--index-url", indexURL,
			"--no-emit-index-url",
		}
		uvCmd := exec.Command(uvPath, uvArgs...)
		uvCmd.Stdout = os.Stdout
		uvCmd.Stderr = os.Stderr
		if err := uvCmd.Run(); err != nil {
			fmt.Printf("Warning: uv resolution failed: %v. Falling back to input as is.\n", err)
			resolvedReq = reqFile
		}
	} else {
		fmt.Println("uv not found or not on Linux. Skipping explicit resolution step.")
	}
	return resolvedReq, nil
}
