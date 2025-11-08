package temingo

import (
	"fmt"
	"path/filepath"
	"strings"
)

// getAbsolutePathsAndRel returns the absolute paths for inputDir and outputDir,
// and the relative path from inputDir to outputDir.
// This helper is used to avoid redundant filepath.Abs calls.
func getAbsolutePathsAndRel(inputDir string, outputDir string) (absInputDir string, absOutputDir string, rel string, err error) {
	absOutputDir, err = filepath.Abs(outputDir)
	if err != nil {
		return "", "", "", fmt.Errorf("error getting absolute path for output directory: %w", err)
	}
	absInputDir, err = filepath.Abs(inputDir)
	if err != nil {
		return "", "", "", fmt.Errorf("error getting absolute path for input directory: %w", err)
	}

	rel, err = filepath.Rel(absInputDir, absOutputDir)
	if err != nil {
		return "", "", "", fmt.Errorf("error determining relationship between input and output directories: %w", err)
	}

	return absInputDir, absOutputDir, rel, nil
}

// GetOutputIgnorePath calculates the relative path of outputDir from inputDir if outputDir is inside inputDir.
// Returns the path to add to ignore list, or empty string if outputDir is outside inputDir.
// This is used to prevent processing loops when outputDir is inside inputDir.
func GetOutputIgnorePath(inputDir string, outputDir string) (string, error) {
	_, absOutputDir, rel, err := getAbsolutePathsAndRel(inputDir, outputDir)
	if err != nil {
		return "", err
	}

	// Check if outputDir is inside or equal to inputDir
	// If rel is ".." or starts with "../", outputDir is outside inputDir (safe, no need to ignore)
	// Otherwise, outputDir is inside or equal to inputDir (needs to be ignored)
	if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		// If outputDir equals inputDir (rel == "."), use the directory name instead
		// This handles the case where inputDir == outputDir
		if rel == "." {
			return filepath.Base(absOutputDir), nil
		}

		return rel, nil
	}

	return "", nil
}
