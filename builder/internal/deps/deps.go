package deps

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"builder/internal/config"
	"builder/internal/utils"
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
	if err := utils.ValidateVersion(fromVer); err != nil {
		return nil, fmt.Errorf("invalid fromVer: %w", err)
	}
	if err := utils.ValidateVersion(toVer); err != nil {
		return nil, fmt.Errorf("invalid toVer: %w", err)
	}

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

	// Copy local path dependencies' wheels to targetDir
	if err := copyLocalWheels(resolvedReq, targetDir); err != nil {
		return "", fmt.Errorf("failed to copy local wheels: %w", err)
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
		} else {
			// Post-process resolvedReq to fix relative paths
			if err := postProcessRequirements(resolvedReq, filepath.Dir(reqFile)); err != nil {
				return "", fmt.Errorf("failed to post-process requirements: %w", err)
			}
		}
	} else {
		fmt.Println("uv not found or not on Linux. Skipping explicit resolution step.")
	}
	return resolvedReq, nil
}

func postProcessRequirements(filePath, baseDir string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip comments, empty lines, and command options
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") {
			lines = append(lines, line)
			continue
		}

		// Check if it represents a local path relative to baseDir
		localPath := filepath.Join(baseDir, trimmed)
		if _, err := os.Stat(localPath); err == nil {
			absPath, err := filepath.Abs(localPath)
			if err == nil {
				fmt.Printf("Post-processing: resolved relative path '%s' to absolute path '%s'\n", trimmed, absPath)
				line = absPath
			}
		}
		lines = append(lines, line)
	}

	f.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	// Write back
	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func copyLocalWheels(reqFile, targetDir string) error {
	f, err := os.Open(reqFile)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string
	pathMap := make(map[string]string) // Map from absolute path to package name

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			lines = append(lines, scanner.Text())
			continue
		}

		// If the line is an absolute path on disk, it's a local dependency
		if filepath.IsAbs(line) {
			if info, err := os.Stat(line); err == nil && info.IsDir() {
				fmt.Printf("Detected local path dependency: %s\n", line)

				// Build the wheel using `uv build --wheel` or `python -m build --wheel`
				uvPath, err := exec.LookPath("uv")
				var buildCmd *exec.Cmd
				if err == nil {
					buildCmd = exec.Command(uvPath, "build", "--wheel")
				} else {
					pythonExe := "python"
					if runtime.GOOS != "windows" {
						if _, err := exec.LookPath("python3"); err == nil {
							pythonExe = "python3"
						}
					}
					buildCmd = exec.Command(pythonExe, "-m", "build", "--wheel")
				}
				buildCmd.Dir = line
				buildCmd.Stdout = os.Stdout
				buildCmd.Stderr = os.Stderr
				fmt.Printf("Building wheel in %s...\n", line)
				if err := buildCmd.Run(); err != nil {
					return fmt.Errorf("failed to build wheel in %s: %w", line, err)
				}

				// Find the generated wheel in the `dist` directory
				distDir := filepath.Join(line, "dist")
				files, err := os.ReadDir(distDir)
				if err != nil {
					return fmt.Errorf("failed to read dist dir %s: %w", distDir, err)
				}

				var newestWheel string
				var newestTime int64
				for _, file := range files {
					if !file.IsDir() && strings.HasSuffix(file.Name(), ".whl") {
						filePath := filepath.Join(distDir, file.Name())
						if fileInfo, err := os.Stat(filePath); err == nil {
							if fileInfo.ModTime().UnixNano() > newestTime {
								newestTime = fileInfo.ModTime().UnixNano()
								newestWheel = filePath
							}
						}
					}
				}

				if newestWheel == "" {
					return fmt.Errorf("no wheel found in %s", distDir)
				}

				// Copy the newest wheel to targetDir
				destPath := filepath.Join(targetDir, filepath.Base(newestWheel))
				fmt.Printf("Copying built wheel %s to %s\n", newestWheel, destPath)
				if err := copyFile(newestWheel, destPath); err != nil {
					return fmt.Errorf("failed to copy wheel: %w", err)
				}

				// Derive the package spec (name==version) from wheel filename (e.g. soda_tracking_service-1.0.2-py3-none-any.whl -> soda-tracking-service==1.0.2)
				baseName := filepath.Base(newestWheel)
				parts := strings.Split(baseName, "-")
				if len(parts) > 1 {
					pkgName := strings.ReplaceAll(parts[0], "_", "-")
					version := parts[1]
					pkgSpec := fmt.Sprintf("%s==%s", pkgName, version)
					pathMap[line] = pkgSpec
					fmt.Printf("Mapped local path '%s' to package spec '%s'\n", line, pkgSpec)
				}
			}
		}
		lines = append(lines, scanner.Text())
	}
	f.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	// Rewrite requirements.txt replacing absolute paths with package names
	if len(pathMap) > 0 {
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if pkgName, exists := pathMap[trimmed]; exists {
				lines[i] = pkgName
			}
		}
		// Write the updated requirements.txt back
		err = os.WriteFile(reqFile, []byte(strings.Join(lines, "\n")+"\n"), 0644)
		if err != nil {
			return fmt.Errorf("failed to rewrite requirements file %s: %w", reqFile, err)
		}
		fmt.Printf("Successfully updated %s with standard package names\n", reqFile)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func GetLocalPackageSpec(dir string) (string, error) {
	pyProject := filepath.Join(dir, "pyproject.toml")
	f, err := os.Open(pyProject)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var name, version string
	inProjectSection := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "[project]" {
			inProjectSection = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inProjectSection = false
			continue
		}

		if inProjectSection {
			if strings.HasPrefix(line, "name") {
				rest := strings.TrimSpace(strings.TrimPrefix(line, "name"))
				if strings.HasPrefix(rest, "=") {
					parts := strings.SplitN(rest, "=", 2)
					val := strings.TrimSpace(parts[1])
					if idx := strings.Index(val, "#"); idx != -1 {
						val = strings.TrimSpace(val[:idx])
					}
					name = strings.Trim(val, "\"'")
				}
			}
			if strings.HasPrefix(line, "version") {
				rest := strings.TrimSpace(strings.TrimPrefix(line, "version"))
				if strings.HasPrefix(rest, "=") {
					parts := strings.SplitN(rest, "=", 2)
					val := strings.TrimSpace(parts[1])
					if idx := strings.Index(val, "#"); idx != -1 {
						val = strings.TrimSpace(val[:idx])
					}
					version = strings.Trim(val, "\"'")
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	if name != "" && version != "" {
		return fmt.Sprintf("%s==%s", name, version), nil
	}
	return "", fmt.Errorf("could not parse name/version from %s", pyProject)
}

func ConvertPathsToSpecs(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string
	hasChanges := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "-") {
			if filepath.IsAbs(trimmed) {
				if info, err := os.Stat(trimmed); err == nil && info.IsDir() {
					spec, err := GetLocalPackageSpec(trimmed)
					if err == nil {
						fmt.Printf("Snapshot: converted local path '%s' to spec '%s'\n", trimmed, spec)
						line = spec
						hasChanges = true
					} else {
						fmt.Printf("Warning: failed to get local package spec for %s: %v\n", trimmed, err)
					}
				}
			}
		}
		lines = append(lines, line)
	}

	f.Close()
	if err := scanner.Err(); err != nil {
		return err
	}

	if hasChanges {
		return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	}
	return nil
}
