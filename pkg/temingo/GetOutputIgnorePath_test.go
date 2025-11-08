package temingo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetOutputIgnorePath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		inputDir    string
		outputDir   string
		setup       func() (string, string) // Returns (inputDir, outputDir)
		wantPath    string                  // Expected ignore path (empty string if output is outside)
		wantErr     bool
		errContains string
		description string
	}{
		{
			name: "outputDir inside inputDir",
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(inputDir, "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantPath:    "output",
			wantErr:     false,
			description: "When outputDir is inside inputDir, should return relative path to ignore",
		},
		{
			name: "outputDir equals inputDir",
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "same")
				os.MkdirAll(inputDir, 0755)
				return inputDir, inputDir
			},
			wantPath:    "", // Will be checked separately as it returns filepath.Base
			wantErr:     false,
			description: "When outputDir equals inputDir, should return base name of directory",
		},
		{
			name: "outputDir outside inputDir",
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(tmpDir, "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantPath:    "",
			wantErr:     false,
			description: "When outputDir is outside inputDir, should return empty string",
		},
		{
			name: "Nested outputDir",
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(inputDir, "nested", "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantPath:    filepath.Join("nested", "output"),
			wantErr:     false,
			description: "When outputDir is nested inside inputDir, should return nested relative path",
		},
		{
			name: "outputDir in sibling directory (outside)",
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(tmpDir, "sibling", "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantPath:    "",
			wantErr:     false,
			description: "When outputDir is in sibling directory, should return empty string",
		},
		{
			name: "Paths with trailing slashes",
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input") + string(filepath.Separator)
				outputDir := filepath.Join(tmpDir, "input", "output") + string(filepath.Separator)
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantPath:    "output",
			wantErr:     false,
			description: "Should handle trailing slashes correctly",
		},
		{
			name: "Output in parent directory (outside)",
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "nested", "input")
				outputDir := filepath.Join(tmpDir, "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantPath:    "",
			wantErr:     false,
			description: "When output is in parent directory, should return empty string",
		},
		{
			name: "outputDir in subdirectory of inputDir",
			setup: func() (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(inputDir, "subdir", "output")
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(outputDir, 0755)
				return inputDir, outputDir
			},
			wantPath:    filepath.Join("subdir", "output"),
			wantErr:     false,
			description: "When outputDir is in subdirectory of inputDir, should return correct relative path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.TempDir() automatically cleans up all files/directories created within it
			// when the test completes, even if the test fails or panics
			inputDir, outputDir := tt.setup()

			// Ensure paths end with separator to match actual usage
			if !strings.HasSuffix(inputDir, string(filepath.Separator)) {
				inputDir = inputDir + string(filepath.Separator)
			}
			if !strings.HasSuffix(outputDir, string(filepath.Separator)) {
				outputDir = outputDir + string(filepath.Separator)
			}

			gotPath, err := GetOutputIgnorePath(inputDir, outputDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetOutputIgnorePath() expected error but got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetOutputIgnorePath() error = %v, expected error to contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("GetOutputIgnorePath() unexpected error: %v (%s)", err, tt.description)
				}

				// For the "Output equals input directory" case, we need to handle the base name comparison
				if tt.name == "outputDir equals inputDir" {
					// The function returns filepath.Base(absOutputDir), so we need to check if it matches
					// the base name of the directory
					absOutputDir, _ := filepath.Abs(strings.TrimSuffix(outputDir, string(filepath.Separator)))
					expectedBase := filepath.Base(absOutputDir)
					if gotPath != expectedBase {
						t.Errorf("GetOutputIgnorePath() = %q, want %q (%s)", gotPath, expectedBase, tt.description)
					}
				} else if tt.wantPath != "" {
					// Only check if we have a specific expected path
					if gotPath != tt.wantPath {
						t.Errorf("GetOutputIgnorePath() = %q, want %q (%s)", gotPath, tt.wantPath, tt.description)
					}
				} else {
					// For empty wantPath, just verify it's empty
					if gotPath != "" {
						t.Errorf("GetOutputIgnorePath() = %q, want empty string (%s)", gotPath, tt.description)
					}
				}
			}
		})
	}
}

func TestGetOutputIgnorePath_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Relative paths are converted to absolute", func(t *testing.T) {
		inputDir := filepath.Join(tmpDir, "input")
		outputDir := filepath.Join(inputDir, "output")
		os.MkdirAll(inputDir, 0755)
		os.MkdirAll(outputDir, 0755)

		// Use relative paths
		relInput := "input"
		relOutput := "input/output"

		// Change to tmpDir to make relative paths work
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		gotPath, err := GetOutputIgnorePath(relInput, relOutput)
		if err != nil {
			t.Errorf("GetOutputIgnorePath() unexpected error: %v", err)
		}
		if gotPath != "output" {
			t.Errorf("GetOutputIgnorePath() = %q, want %q", gotPath, "output")
		}
	})

	t.Run("Same directory with different representations", func(t *testing.T) {
		// t.TempDir() automatically cleans up all files/directories created within it
		// when the test completes, even if the test fails or panics
		inputDir := filepath.Join(tmpDir, "testdir")
		os.MkdirAll(inputDir, 0755)

		// Test with and without trailing slashes
		testCases := []struct {
			input  string
			output string
		}{
			{inputDir, inputDir},
			{inputDir + "/", inputDir + "/"},
			{inputDir, inputDir + "/"},
			{inputDir + "/", inputDir},
		}

		for _, tc := range testCases {
			gotPath, err := GetOutputIgnorePath(tc.input, tc.output)
			if err != nil {
				t.Errorf("GetOutputIgnorePath(%q, %q) unexpected error: %v", tc.input, tc.output, err)
				continue
			}

			// Should return the base name of the directory
			absOutputDir, _ := filepath.Abs(tc.output)
			expectedBase := filepath.Base(absOutputDir)
			if gotPath != expectedBase {
				t.Errorf("GetOutputIgnorePath(%q, %q) = %q, want %q", tc.input, tc.output, gotPath, expectedBase)
			}
		}
	})

	t.Run("Deeply nested output directory", func(t *testing.T) {
		// t.TempDir() automatically cleans up all files/directories created within it
		// when the test completes, even if the test fails or panics
		inputDir := filepath.Join(tmpDir, "input")
		outputDir := filepath.Join(inputDir, "level1", "level2", "level3", "output")
		os.MkdirAll(inputDir, 0755)
		os.MkdirAll(outputDir, 0755)

		gotPath, err := GetOutputIgnorePath(inputDir+"/", outputDir+"/")
		if err != nil {
			t.Errorf("GetOutputIgnorePath() unexpected error: %v", err)
		}

		expectedPath := filepath.Join("level1", "level2", "level3", "output")
		if gotPath != expectedPath {
			t.Errorf("GetOutputIgnorePath() = %q, want %q", gotPath, expectedPath)
		}
	})
}
