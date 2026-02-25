package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var versionRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// ValidateVersion checks if the version string is safe for use in filenames.
// It allows alphanumeric characters, underscores, dots, and hyphens.
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	if !versionRegex.MatchString(version) {
		return fmt.Errorf("version contains invalid characters: %s", version)
	}
	return nil
}

// SanitizeNSISPath checks for dangerous characters in a path intended for NSIS scripts.
// It returns a quoted string safe for use in NSIS commands like !addplugindir.
func SanitizeNSISPath(path string) (string, error) {
	if strings.ContainsAny(path, "\n\r") {
		return "", fmt.Errorf("path contains newlines: %s", path)
	}

	// NSIS supports ", ', and ` as delimiters.
	// We try to find one that isn't in the path.
	if !strings.Contains(path, "\"") {
		return "\"" + path + "\"", nil
	}
	if !strings.Contains(path, "'") {
		return "'" + path + "'", nil
	}
	if !strings.Contains(path, "`") {
		return "`" + path + "`", nil
	}

	return "", fmt.Errorf("path contains all possible NSIS delimiters (\", ', `): %s", path)
}
