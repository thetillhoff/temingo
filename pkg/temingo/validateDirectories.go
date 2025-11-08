package temingo

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// validateDirectories checks that inputDir exists and is a directory,
// and that outputDir exists and is a directory (or creates it if it doesn't exist).
// If inputDir == outputDir and noDeleteOutputDir is false, returns an error since this would delete the input directory.
// If outputDir is inside inputDir (but not equal), it will be added to the ignore list at runtime to prevent loops.
// Returns the ignore path (or empty string if output is outside input) and any error.
// This function is used by Render() which needs the input/output comparison logic.
func validateDirectories(inputDir string, outputDir string, noDeleteOutputDir bool) (string, error) {
	// Validate inputDir
	info, err := os.Stat(inputDir)
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
	info, err = os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Get permissions from input directory to preserve them
			inputInfo, err := os.Stat(inputDir)
			if err != nil {
				return "", fmt.Errorf("error getting input directory info: %w", err)
			}
			// Create output directory with same permissions as input directory
			err = os.MkdirAll(outputDir, inputInfo.Mode().Perm())
			if err != nil {
				return "", fmt.Errorf("error creating output directory %s: %w", outputDir, err)
			}
			log.Printf("Created output directory: %s", outputDir)
		} else {
			return "", fmt.Errorf("error accessing output directory %s: %w", outputDir, err)
		}
	} else if !info.IsDir() {
		return "", fmt.Errorf("output directory is not a directory: %s", outputDir)
	}

	// Check if inputDir == outputDir using the helper function
	// We use getAbsolutePathsAndRel to avoid redundant filepath.Abs calls
	_, absOutputDir, rel, err := getAbsolutePathsAndRel(inputDir, outputDir)
	if err != nil {
		return "", err
	}

	// Calculate the ignore path (same logic as GetOutputIgnorePath)
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
