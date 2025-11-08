package temingo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		outputDir   string
		setup       func() string // Returns outputDir
		wantErr     bool
		errContains string
		description string
	}{
		{
			name: "Output directory does not exist - should be created",
			setup: func() string {
				outputDir := filepath.Join(tmpDir, "output")
				// Don't create it - let ensureOutputDirectory create it
				return outputDir
			},
			wantErr:     false,
			description: "When outputDir does not exist, it should be created with default permissions (0755)",
		},
		{
			name: "Output directory already exists - should succeed",
			setup: func() string {
				outputDir := filepath.Join(tmpDir, "output")
				os.MkdirAll(outputDir, 0755)
				return outputDir
			},
			wantErr:     false,
			description: "When outputDir already exists, it should succeed without error",
		},
		{
			name: "Output is a file, not a directory - should fail",
			setup: func() string {
				outputFile := filepath.Join(tmpDir, "output")
				os.WriteFile(outputFile, []byte("test"), 0644)
				return outputFile
			},
			wantErr:     true,
			errContains: "output directory is not a directory",
			description: "When outputDir exists but is a file, it should return an error",
		},
		{
			name: "Nested output directory does not exist - should be created",
			setup: func() string {
				outputDir := filepath.Join(tmpDir, "nested", "output", "dir")
				// Don't create it - let ensureOutputDirectory create it
				return outputDir
			},
			wantErr:     false,
			description: "When nested outputDir does not exist, it should be created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := tt.setup()

			// Ensure path ends with separator to match actual usage
			if !strings.HasSuffix(outputDir, string(filepath.Separator)) {
				outputDir = outputDir + string(filepath.Separator)
			}

			err := ensureOutputDirectory(outputDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ensureOutputDirectory() expected error but got nil (%s)", tt.description)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ensureOutputDirectory() error = %v, expected error to contain %q (%s)", err, tt.errContains, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("ensureOutputDirectory() unexpected error: %v (%s)", err, tt.description)
				}

				// Verify the directory was created/exists
				info, err := os.Stat(strings.TrimSuffix(outputDir, string(filepath.Separator)))
				if err != nil {
					t.Errorf("ensureOutputDirectory() directory should exist but stat failed: %v (%s)", err, tt.description)
				} else if !info.IsDir() {
					t.Errorf("ensureOutputDirectory() path exists but is not a directory (%s)", tt.description)
				}
			}
		})
	}
}

func TestEnsureOutputDirectory_Permissions(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Created directory should have default permissions (0755)", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "output")

		err := ensureOutputDirectory(outputDir + string(filepath.Separator))
		if err != nil {
			t.Fatalf("ensureOutputDirectory() unexpected error: %v", err)
		}

		info, err := os.Stat(outputDir)
		if err != nil {
			t.Fatalf("Failed to stat output directory: %v", err)
		}

		actualPerm := info.Mode().Perm()
		expectedPerm := os.FileMode(0755)
		if actualPerm != expectedPerm {
			t.Errorf("Output directory permissions = %o, want %o", actualPerm, expectedPerm)
		}
	})
}
