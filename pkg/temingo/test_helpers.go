package temingo

import (
	"os"
	"path/filepath"

	"github.com/thetillhoff/fileIO"
)

// setupTestProjectFromInitFiles sets up a test project by copying files from InitFiles
// This allows tests to use the same example projects that InitProject uses
// Returns inputDir, outputDir, and any error
func setupTestProjectFromInitFiles(tmpDir string, projectType string) (string, string, error) {
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	// Create directories
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", err
	}

	// Get example project files using the same logic as InitProject
	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.TemingoignorePath = filepath.Join(tmpDir, ".temingoignore")

	files, err := engine.getExampleProjectFiles(projectType)
	if err != nil {
		return "", "", err
	}

	// Write files to disk
	// The paths returned by getExampleProjectFiles are already adjusted to use engine.InputDir
	// and engine.TemingoignorePath, so we can write them directly
	for path, content := range files {
		// Ensure parent directory exists
		parentDir := filepath.Dir(path)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return "", "", err
		}

		if err := fileIO.WriteFile(path, content); err != nil {
			return "", "", err
		}
	}

	return inputDir, outputDir, nil
}

// setupTestProjectFromInitFilesWithEngine sets up a test project and returns an engine configured for it
// This is useful for tests that need to render the project
func setupTestProjectFromInitFilesWithEngine(tmpDir string, projectType string, values map[string]string) (*Engine, string, string, error) {
	inputDir, outputDir, err := setupTestProjectFromInitFiles(tmpDir, projectType)
	if err != nil {
		return nil, "", "", err
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.TemingoignorePath = filepath.Join(tmpDir, ".temingoignore")
	if values != nil {
		engine.Values = values
	}
	engine.Verbose = false

	return &engine, inputDir, outputDir, nil
}
