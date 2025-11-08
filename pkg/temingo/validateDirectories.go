package temingo

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// validateDirectories checks that inputDir exists and is a directory,
// and that outputDir exists and is a directory (or creates it if it doesn't exist).
// If inputDir == outputDir and noDeleteOutputDir is false, returns an error since this would delete the input directory.
// If outputDir is inside inputDir (but not equal), it will be added to the ignore list at runtime to prevent loops.
// Returns the ignore path (or empty string if output is outside input) and any error.
func validateDirectories(inputDir string, outputDir string, noDeleteOutputDir bool, logger *slog.Logger) (string, error) {
	// Initialize logger if not provided (use default logger as fallback)
	if logger == nil {
		logger = slog.Default()
	}

	// Validate inputDir
	// Handle trailing separator - if path ends with separator, try without it first
	inputDirToCheck := strings.TrimSuffix(inputDir, string(filepath.Separator))
	if inputDirToCheck == "" {
		inputDirToCheck = inputDir
	}

	// Normalize the path to ensure consistent handling
	inputDirToCheck = filepath.Clean(inputDirToCheck)
	info, err := os.Stat(inputDirToCheck)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("input directory does not exist: %s", inputDir)
		}
		return "", fmt.Errorf("error accessing input directory %s: %w", inputDir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("input directory is not a directory: %s", inputDir)
	}

	// Validate or create outputDir
	// Handle trailing separator - if path ends with separator, try without it first
	outputDirToCheck := strings.TrimSuffix(outputDir, string(filepath.Separator))
	if outputDirToCheck == "" {
		outputDirToCheck = outputDir
	}

	// Normalize the path to ensure consistent handling
	outputDirToCheck = filepath.Clean(outputDirToCheck)
	info, err = os.Stat(outputDirToCheck)
	if err != nil {
		if os.IsNotExist(err) {
			// Get permissions from input directory to preserve them
			inputInfo, err := os.Stat(inputDirToCheck)
			if err != nil {
				return "", fmt.Errorf("error getting input directory info: %w", err)
			}

			// Create output directory with same permissions as input directory
			// Use the cleaned path without separator for MkdirAll
			err = os.MkdirAll(outputDirToCheck, inputInfo.Mode().Perm())
			if err != nil {
				return "", fmt.Errorf("error creating output directory %s: %w", outputDir, err)
			}

			// Use Chmod to ensure exact permissions (MkdirAll may be affected by umask)
			err = os.Chmod(outputDirToCheck, inputInfo.Mode().Perm())
			if err != nil {
				return "", fmt.Errorf("error setting output directory permissions %s: %w", outputDir, err)
			}
			logger.Info("Created output directory", "path", outputDir)
		} else {
			return "", fmt.Errorf("error accessing output directory %s: %w", outputDir, err)
		}
	} else if !info.IsDir() {
		return "", fmt.Errorf("output directory is not a directory: %s", outputDir)
	}

	// Check if inputDir == outputDir using the helper function
	// We use getAbsolutePathsAndRel to avoid redundant filepath.Abs calls
	_, absOutputDir, rel, err := getAbsolutePathsAndRelativePath(inputDir, outputDir)
	if err != nil {
		return "", err
	}

	// Calculate the ignore path
	var outputIgnorePath string
	if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		// If outputDir equals inputDir (rel == "."), use the directory name instead
		if rel == "." {
			outputIgnorePath = filepath.Base(absOutputDir)
		} else {
			outputIgnorePath = rel
		}
	}

	// Check if input and output are the same directory (rel == "."), check if --noDeleteOutputDir is set
	if rel == "." {
		if !noDeleteOutputDir {
			return "", fmt.Errorf("input directory cannot equal output directory when --noDeleteOutputDir is not set (this would delete the input directory)")
		}
		// If --noDeleteOutputDir is set, it's allowed (output will be added to ignore list in Render.go)
	}
	// If output is inside but not equal, it's allowed (will be added to ignore list in Render.go)

	return outputIgnorePath, nil
}
