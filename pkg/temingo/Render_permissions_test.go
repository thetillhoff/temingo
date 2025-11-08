package temingo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRender_PreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		inputPerm   os.FileMode
		description string
	}{
		{
			name:        "Template files should have same permissions as input directory (0755)",
			inputPerm:   0755,
			description: "Rendered template files should preserve input directory permissions",
		},
		{
			name:        "Template files should have same permissions as input directory (0700)",
			inputPerm:   0700,
			description: "Rendered template files should preserve restrictive input directory permissions",
		},
		{
			name:        "Template files should have same permissions as input directory (0777)",
			inputPerm:   0777,
			description: "Rendered template files should preserve permissive input directory permissions",
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
			err = os.Chmod(inputDir, tt.inputPerm)
			if err != nil {
				t.Fatalf("Failed to set input directory permissions: %v", err)
			}

			// Create a simple template file
			templateFile := filepath.Join(inputDir, "index.template.html")
			err = os.WriteFile(templateFile, []byte("Hello {{ .path }}"), 0644)
			if err != nil {
				t.Fatalf("Failed to create template file: %v", err)
			}

			// Create engine and render
			engine := DefaultEngine()
			engine.InputDir = inputDir + string(filepath.Separator)
			engine.OutputDir = outputDir + string(filepath.Separator)
			engine.Verbose = false

			err = engine.Render()
			if err != nil {
				t.Fatalf("Render() unexpected error: %v", err)
			}

			// Check output file permissions
			outputFile := filepath.Join(outputDir, "index.html")
			info, err := os.Stat(outputFile)
			if err != nil {
				t.Fatalf("Failed to stat output file: %v", err)
			}

			actualPerm := info.Mode().Perm()
			expectedPerm := tt.inputPerm
			if actualPerm != expectedPerm {
				t.Errorf("Output file permissions = %o, want %o (%s)", actualPerm, expectedPerm, tt.description)
			}

			// Check output directory permissions (should also match input)
			outputDirInfo, err := os.Stat(outputDir)
			if err != nil {
				t.Fatalf("Failed to stat output directory: %v", err)
			}
			outputDirPerm := outputDirInfo.Mode().Perm()
			if outputDirPerm != expectedPerm {
				t.Errorf("Output directory permissions = %o, want %o", outputDirPerm, expectedPerm)
			}
		})
	}
}
