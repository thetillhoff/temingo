package temingo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupTestProjectFromInitFiles(t *testing.T) {
	tests := []struct {
		name        string
		projectType string
		description string
	}{
		{
			name:        "Example project",
			projectType: "example",
			description: "Should set up example project from InitFiles",
		},
		{
			name:        "Test project",
			projectType: "test",
			description: "Should set up test project from InitFiles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new tmpDir for each test case to ensure isolation
			tmpDir := t.TempDir()

			inputDir, outputDir, err := setupTestProjectFromInitFiles(tmpDir, tt.projectType)
			if err != nil {
				t.Fatalf("setupTestProjectFromInitFiles() failed: %v", err)
			}

			// Verify directories were created
			if _, err := os.Stat(inputDir); os.IsNotExist(err) {
				t.Errorf("setupTestProjectFromInitFiles() should have created inputDir %q", inputDir)
			}
			if _, err := os.Stat(outputDir); os.IsNotExist(err) {
				t.Errorf("setupTestProjectFromInitFiles() should have created outputDir %q", outputDir)
			}

			// Verify that files from InitFiles were copied
			// Check for .temingoignore
			ignoreFile := filepath.Join(tmpDir, ".temingoignore")
			if _, err := os.Stat(ignoreFile); os.IsNotExist(err) {
				t.Errorf("setupTestProjectFromInitFiles() should have created .temingoignore file")
			}

			// Check for src directory
			srcDir := filepath.Join(inputDir)
			if _, err := os.Stat(srcDir); os.IsNotExist(err) {
				t.Errorf("setupTestProjectFromInitFiles() should have created src directory")
			}

			// Verify at least one template file exists
			indexFile := filepath.Join(inputDir, "index.template.html")
			if _, err := os.Stat(indexFile); os.IsNotExist(err) {
				// For test project, check if there are other files
				entries, err := os.ReadDir(inputDir)
				if err != nil {
					t.Fatalf("Failed to read input directory: %v", err)
				}
				if len(entries) == 0 {
					t.Errorf("setupTestProjectFromInitFiles() should have created files in inputDir")
				}
			}
		})
	}
}

func TestSetupTestProjectFromInitFilesWithEngine(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	values := map[string]string{
		"testKey": "testValue",
	}

	engine, inputDir, outputDir, err := setupTestProjectFromInitFilesWithEngine(tmpDir, "example", values)
	if err != nil {
		t.Fatalf("setupTestProjectFromInitFilesWithEngine() failed: %v", err)
	}

	// Verify engine is configured correctly
	if engine.InputDir != inputDir+string(filepath.Separator) {
		t.Errorf("engine.InputDir = %q, want %q", engine.InputDir, inputDir+string(filepath.Separator))
	}
	if engine.OutputDir != outputDir+string(filepath.Separator) {
		t.Errorf("engine.OutputDir = %q, want %q", engine.OutputDir, outputDir+string(filepath.Separator))
	}
	if engine.Values["testKey"] != "testValue" {
		t.Errorf("engine.Values[\"testKey\"] = %q, want %q", engine.Values["testKey"], "testValue")
	}

	// Verify directories exist
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		t.Errorf("setupTestProjectFromInitFilesWithEngine() should have created inputDir")
	}
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Errorf("setupTestProjectFromInitFilesWithEngine() should have created outputDir")
	}
}

