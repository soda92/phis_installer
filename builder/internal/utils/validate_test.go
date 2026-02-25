package utils

import "testing"

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"v1.0", true},
		{"1.0.0-beta", true},
		{"1_0_0", true},
		{"../1.0.0", false},
		{"1.0.0/../", false},
		{"foo/bar", false},
		{"foo\bar", false},
		{"", false},
		{"  ", false},
	}

	for _, tt := range tests {
		err := ValidateVersion(tt.version)
		if tt.valid && err != nil {
			t.Errorf("ValidateVersion(%q) returned error: %v, expected valid", tt.version, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ValidateVersion(%q) returned nil, expected error", tt.version)
		}
	}
}
