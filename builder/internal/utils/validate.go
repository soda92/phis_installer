package utils

import (
	"fmt"
	"regexp"
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
