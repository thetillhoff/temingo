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

	// Normalize all paths to be relative to tmpDir
	// If a path is absolute, make it relative to tmpDir
	// If a path is relative, keep it as-is
	normalizedFiles := make(map[string][]byte)
	for path, content := range files {
		var relPath string
		if filepath.IsAbs(path) {
			// Path is absolute - make it relative to tmpDir
			relPath, err = filepath.Rel(tmpDir, path)
			if err != nil {
				return "", "", err
			}
		} else {
			// Path is already relative, use it as-is
			relPath = path
		}
		normalizedFiles[relPath] = content
	}
	files = normalizedFiles

	// Write files to disk - join relative paths with tmpDir to get absolute paths
	for relPath, content := range files {
		absPath := filepath.Join(tmpDir, relPath)
		// Ensure parent directory exists
		parentDir := filepath.Dir(absPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return "", "", err
		}

		if err := fileIO.WriteFile(absPath, content); err != nil {
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
