package temingo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetAbsolutePathsAndRelativePath(t *testing.T) {
	tests := []struct {
		name           string
		inputDir       string
		outputDir      string
		setup          func(t *testing.T, tmpDir string) (string, string) // Returns (inputDir, outputDir)
		wantRel        string                                             // Expected relative path
		wantErr        bool
		errContains    string
		description    string
		checkAbsInput  bool // Whether to verify absolute input path
		checkAbsOutput bool // Whether to verify absolute output path
	}{
		{
			name: "outputDir inside inputDir",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(inputDir, "output")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatalf("Failed to create output directory: %v", err)
				}
				return inputDir, outputDir
			},
			wantRel:        "output",
			wantErr:        false,
			checkAbsInput:  true,
			checkAbsOutput: true,
			description:    "When outputDir is inside inputDir, should return relative path",
		},
		{
			name: "outputDir equals inputDir",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				inputDir := filepath.Join(tmpDir, "same")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
				return inputDir, inputDir
			},
			wantRel:        ".",
			wantErr:        false,
			checkAbsInput:  true,
			checkAbsOutput: true,
			description:    "When outputDir equals inputDir, should return '.'",
		},
		{
			name: "outputDir outside inputDir",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(tmpDir, "output")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatalf("Failed to create output directory: %v", err)
				}
				return inputDir, outputDir
			},
			wantRel:        "../output",
			wantErr:        false,
			checkAbsInput:  true,
			checkAbsOutput: true,
			description:    "When outputDir is outside inputDir, should return relative path with '..'",
		},
		{
			name: "Nested outputDir",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(inputDir, "nested", "output")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatalf("Failed to create output directory: %v", err)
				}
				return inputDir, outputDir
			},
			wantRel:        filepath.Join("nested", "output"),
			wantErr:        false,
			checkAbsInput:  true,
			checkAbsOutput: true,
			description:    "When outputDir is nested inside inputDir, should return nested relative path",
		},
		{
			name: "outputDir in sibling directory",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(tmpDir, "sibling", "output")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatalf("Failed to create output directory: %v", err)
				}
				return inputDir, outputDir
			},
			wantRel:        filepath.Join("..", "sibling", "output"),
			wantErr:        false,
			checkAbsInput:  true,
			checkAbsOutput: true,
			description:    "When outputDir is in sibling directory, should return relative path with '..'",
		},
		{
			name: "Paths with trailing slashes",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				inputDir := filepath.Join(tmpDir, "input") + string(filepath.Separator)
				outputDir := filepath.Join(tmpDir, "input", "output") + string(filepath.Separator)
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatalf("Failed to create output directory: %v", err)
				}
				return inputDir, outputDir
			},
			wantRel:        "output",
			wantErr:        false,
			checkAbsInput:  true,
			checkAbsOutput: true,
			description:    "Should handle trailing slashes correctly",
		},
		{
			name: "Output in parent directory",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				inputDir := filepath.Join(tmpDir, "nested", "input")
				outputDir := filepath.Join(tmpDir, "output")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatalf("Failed to create output directory: %v", err)
				}
				return inputDir, outputDir
			},
			wantRel:        filepath.Join("..", "..", "output"),
			wantErr:        false,
			checkAbsInput:  true,
			checkAbsOutput: true,
			description:    "When output is in parent directory, should return relative path with '..'",
		},
		{
			name: "outputDir in subdirectory of inputDir",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				inputDir := filepath.Join(tmpDir, "input")
				outputDir := filepath.Join(inputDir, "subdir", "output")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatalf("Failed to create output directory: %v", err)
				}
				return inputDir, outputDir
			},
			wantRel:        filepath.Join("subdir", "output"),
			wantErr:        false,
			checkAbsInput:  true,
			checkAbsOutput: true,
			description:    "When outputDir is in subdirectory of inputDir, should return correct relative path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputDir, outputDir := tt.setup(t, tmpDir)

			gotAbsInput, gotAbsOutput, gotRel, err := getAbsolutePathsAndRelativePath(inputDir, outputDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("getAbsolutePathsAndRelativePath() expected error but got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("getAbsolutePathsAndRelativePath() error = %v, expected error to contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("getAbsolutePathsAndRelativePath() unexpected error: %v (%s)", err, tt.description)
				}

				if tt.checkAbsInput {
					expectedAbsInput, _ := filepath.Abs(inputDir)
					if gotAbsInput != expectedAbsInput {
						t.Errorf("getAbsolutePathsAndRelativePath() absInputDir = %q, want %q (%s)", gotAbsInput, expectedAbsInput, tt.description)
					}
				}

				if tt.checkAbsOutput {
					expectedAbsOutput, _ := filepath.Abs(outputDir)
					if gotAbsOutput != expectedAbsOutput {
						t.Errorf("getAbsolutePathsAndRelativePath() absOutputDir = %q, want %q (%s)", gotAbsOutput, expectedAbsOutput, tt.description)
					}
				}

				if gotRel != tt.wantRel {
					t.Errorf("getAbsolutePathsAndRelativePath() rel = %q, want %q (%s)", gotRel, tt.wantRel, tt.description)
				}
			}
		})
	}
}

func TestGetAbsolutePathsAndRelativePath_EdgeCases(t *testing.T) {
	t.Run("Relative paths are converted to absolute", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputDir := filepath.Join(tmpDir, "input")
		outputDir := filepath.Join(inputDir, "output")
		if err := os.MkdirAll(inputDir, 0755); err != nil {
			t.Fatalf("Failed to create input directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		gotAbsInput, gotAbsOutput, gotRel, err := getAbsolutePathsAndRelativePath(inputDir, outputDir)
		if err != nil {
			t.Errorf("getAbsolutePathsAndRelativePath() unexpected error: %v", err)
		}

		expectedAbsInput, _ := filepath.Abs(inputDir)
		expectedAbsOutput, _ := filepath.Abs(outputDir)
		if gotAbsInput != expectedAbsInput {
			t.Errorf("getAbsolutePathsAndRelativePath() absInputDir = %q, want %q", gotAbsInput, expectedAbsInput)
		}
		if gotAbsOutput != expectedAbsOutput {
			t.Errorf("getAbsolutePathsAndRelativePath() absOutputDir = %q, want %q", gotAbsOutput, expectedAbsOutput)
		}
		if gotRel != "output" {
			t.Errorf("getAbsolutePathsAndRelativePath() rel = %q, want %q", gotRel, "output")
		}
	})

	t.Run("Same directory with different representations", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputDir := filepath.Join(tmpDir, "testdir")
		if err := os.MkdirAll(inputDir, 0755); err != nil {
			t.Fatalf("Failed to create input directory: %v", err)
		}

		testCases := []struct {
			input  string
			output string
		}{
			{inputDir, inputDir},
			{inputDir + string(filepath.Separator), inputDir + string(filepath.Separator)},
			{inputDir, inputDir + string(filepath.Separator)},
			{inputDir + string(filepath.Separator), inputDir},
		}

		for _, tc := range testCases {
			gotAbsInput, gotAbsOutput, gotRel, err := getAbsolutePathsAndRelativePath(tc.input, tc.output)
			if err != nil {
				t.Errorf("getAbsolutePathsAndRelativePath(%q, %q) unexpected error: %v", tc.input, tc.output, err)
				continue
			}

			expectedAbsInput, _ := filepath.Abs(tc.input)
			expectedAbsOutput, _ := filepath.Abs(tc.output)
			if gotAbsInput != expectedAbsInput {
				t.Errorf("getAbsolutePathsAndRelativePath(%q, %q) absInputDir = %q, want %q", tc.input, tc.output, gotAbsInput, expectedAbsInput)
			}
			if gotAbsOutput != expectedAbsOutput {
				t.Errorf("getAbsolutePathsAndRelativePath(%q, %q) absOutputDir = %q, want %q", tc.input, tc.output, gotAbsOutput, expectedAbsOutput)
			}
			if gotRel != "." {
				t.Errorf("getAbsolutePathsAndRelativePath(%q, %q) rel = %q, want %q", tc.input, tc.output, gotRel, ".")
			}
		}
	})

	t.Run("Deeply nested output directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputDir := filepath.Join(tmpDir, "input")
		outputDir := filepath.Join(inputDir, "level1", "level2", "level3", "output")
		if err := os.MkdirAll(inputDir, 0755); err != nil {
			t.Fatalf("Failed to create input directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		gotAbsInput, gotAbsOutput, gotRel, err := getAbsolutePathsAndRelativePath(inputDir, outputDir)
		if err != nil {
			t.Errorf("getAbsolutePathsAndRelativePath() unexpected error: %v", err)
		}

		expectedAbsInput, _ := filepath.Abs(inputDir)
		expectedAbsOutput, _ := filepath.Abs(outputDir)
		if gotAbsInput != expectedAbsInput {
			t.Errorf("getAbsolutePathsAndRelativePath() absInputDir = %q, want %q", gotAbsInput, expectedAbsInput)
		}
		if gotAbsOutput != expectedAbsOutput {
			t.Errorf("getAbsolutePathsAndRelativePath() absOutputDir = %q, want %q", gotAbsOutput, expectedAbsOutput)
		}

		expectedRel := filepath.Join("level1", "level2", "level3", "output")
		if gotRel != expectedRel {
			t.Errorf("getAbsolutePathsAndRelativePath() rel = %q, want %q", gotRel, expectedRel)
		}
	})
}
