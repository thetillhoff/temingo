package temingo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateDirectories(t *testing.T) {
	// Create temporary directories for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name              string
		inputDir          string
		outputDir         string
		noDeleteOutputDir bool
		setup             func() (string, string) // Returns (inputDir, outputDir)
		cleanup           func(string, string)
		wantErr           bool
		errContains       string
		wantIgnorePath    string // Expected ignore path (empty if output is outside input)
	}{
		{
			name:              "Valid separate directories",
			noDeleteOutputDir: false,
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(tmpDir, "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantErr:        false,
			wantIgnorePath: "", // Output is outside input, so no ignore path
		},
		{
			name:              "Input directory does not exist",
			noDeleteOutputDir: false,
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "nonexistent")
				outputDir := filepath.Join(tmpDir, "output")
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantErr:     true,
			errContains: "input directory does not exist",
		},
		{
			name:              "Output directory does not exist - should be created",
			noDeleteOutputDir: false,
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(tmpDir, "output")
				os.MkdirAll(inputDir, 0755)
				return inputDir, outputDir
			},
			wantErr: false,
		},
		{
			name:              "Input equals output without --noDeleteOutputDir - should fail",
			noDeleteOutputDir: false,
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "same")
				os.MkdirAll(inputDir, 0755)
				return inputDir, inputDir
			},
			wantErr:     true,
			errContains: "input directory cannot equal output directory when --noDeleteOutputDir is not set",
		},
		{
			name:              "Input equals output with --noDeleteOutputDir - should succeed",
			noDeleteOutputDir: true,
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "same")
				os.MkdirAll(inputDir, 0755)
				return inputDir, inputDir
			},
			wantErr:        false,
			wantIgnorePath: "", // Will be checked separately as it returns filepath.Base
		},
		{
			name:              "Output inside input - should succeed (will be ignored at runtime)",
			noDeleteOutputDir: false,
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(inputDir, "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantErr:        false,
			wantIgnorePath: "output", // Output is inside input, should return relative path
		},
		{
			name:              "Input is a file, not a directory",
			noDeleteOutputDir: false,
			setup: func() (string, string) {
				inputFile := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(tmpDir, "output")
				os.WriteFile(inputFile, []byte("test"), 0644)
				os.MkdirAll(outputDir, 0755)
				return inputFile, outputDir
			},
			wantErr:     true,
			errContains: "input directory is not a directory",
		},
		{
			name:              "Output is a file, not a directory",
			noDeleteOutputDir: false,
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputFile := filepath.Join(tmpDir, "output")
				os.MkdirAll(inputDir, 0755)
				os.WriteFile(outputFile, []byte("test"), 0644)
				return inputDir, outputFile
			},
			wantErr:     true,
			errContains: "output directory is not a directory",
		},
		{
			name:              "Output outside input - should succeed",
			noDeleteOutputDir: false,
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(tmpDir, "..", "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantErr:        false,
			wantIgnorePath: "", // Output is outside input, so no ignore path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputDir, outputDir := tt.setup()

			// Ensure paths end with separator to match actual usage
			if !strings.HasSuffix(inputDir, string(filepath.Separator)) {
				inputDir = inputDir + string(filepath.Separator)
			}
			if !strings.HasSuffix(outputDir, string(filepath.Separator)) {
				outputDir = outputDir + string(filepath.Separator)
			}

			gotIgnorePath, err := validateDirectories(inputDir, outputDir, tt.noDeleteOutputDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateDirectories() expected error but got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateDirectories() error = %v, expected error to contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDirectories() unexpected error: %v", err)
				}

				// Check ignore path for "Input equals output" case
				if tt.name == "Input equals output with --noDeleteOutputDir - should succeed" {
					// The function returns filepath.Base(absOutputDir), so we need to check if it matches
					absOutputDir, _ := filepath.Abs(strings.TrimSuffix(outputDir, string(filepath.Separator)))
					expectedBase := filepath.Base(absOutputDir)
					if gotIgnorePath != expectedBase {
						t.Errorf("ValidateDirectories() ignorePath = %q, want %q", gotIgnorePath, expectedBase)
					}
				} else if tt.wantIgnorePath != "" {
					// For other cases, check the expected ignore path
					if gotIgnorePath != tt.wantIgnorePath {
						t.Errorf("ValidateDirectories() ignorePath = %q, want %q", gotIgnorePath, tt.wantIgnorePath)
					}
				} else {
					// For empty wantIgnorePath, just verify it's empty
					if gotIgnorePath != "" {
						t.Errorf("ValidateDirectories() ignorePath = %q, want empty string", gotIgnorePath)
					}
				}
			}
		})
	}
}

func TestValidateDirectories_Permissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		inputPerm    os.FileMode
		outputExists bool
		outputPerm   os.FileMode
		expectedPerm os.FileMode
		description  string
	}{
		{
			name:         "OutputDir created - should match inputDir permissions (0755)",
			inputPerm:    0755,
			outputExists: false,
			expectedPerm: 0755,
			description:  "When outputDir is created, it should have the same permissions as inputDir",
		},
		{
			name:         "OutputDir created - should match inputDir permissions (0700)",
			inputPerm:    0700,
			outputExists: false,
			expectedPerm: 0700,
			description:  "When outputDir is created with restrictive permissions, it should preserve them",
		},
		{
			name:         "OutputDir created - should match inputDir permissions (0777)",
			inputPerm:    0777,
			outputExists: false,
			expectedPerm: 0777,
			description:  "When outputDir is created with permissive permissions, it should preserve them",
		},
		{
			name:         "OutputDir already exists - should not change permissions",
			inputPerm:    0755,
			outputExists: true,
			outputPerm:   0700,
			expectedPerm: 0700,
			description:  "When outputDir already exists, its existing permissions should remain unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputDir := filepath.Join(tmpDir, "input")
			outputDir := filepath.Join(tmpDir, "output")

			// Create input directory with specified permissions
			err := os.MkdirAll(inputDir, tt.inputPerm)
			if err != nil {
				t.Fatalf("Failed to create input directory: %v", err)
			}

			// Set input directory permissions explicitly
			err = os.Chmod(inputDir, tt.inputPerm)
			if err != nil {
				t.Fatalf("Failed to set input directory permissions: %v", err)
			}

			// Create output directory if it should exist
			if tt.outputExists {
				err := os.MkdirAll(outputDir, tt.outputPerm)
				if err != nil {
					t.Fatalf("Failed to create output directory: %v", err)
				}
				err = os.Chmod(outputDir, tt.outputPerm)
				if err != nil {
					t.Fatalf("Failed to set output directory permissions: %v", err)
				}
			}

			// Ensure paths end with separator
			if !strings.HasSuffix(inputDir, string(filepath.Separator)) {
				inputDir = inputDir + string(filepath.Separator)
			}
			if !strings.HasSuffix(outputDir, string(filepath.Separator)) {
				outputDir = outputDir + string(filepath.Separator)
			}

			// Run validation
			_, err = validateDirectories(inputDir, outputDir, false)
			if err != nil {
				t.Fatalf("ValidateDirectories() unexpected error: %v", err)
			}

			// Check output directory permissions
			info, err := os.Stat(outputDir)
			if err != nil {
				t.Fatalf("Failed to stat output directory: %v", err)
			}

			actualPerm := info.Mode().Perm()
			if actualPerm != tt.expectedPerm {
				t.Errorf("Output directory permissions = %o, want %o (%s)", actualPerm, tt.expectedPerm, tt.description)
			}
		})
	}
}
