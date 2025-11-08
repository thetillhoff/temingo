package temingo

import (
	"fmt"
	"path/filepath"
)

// getAbsolutePathsAndRel returns the absolute paths for inputDir and outputDir,
// and the relative path from inputDir to outputDir.
func getAbsolutePathsAndRelativePath(inputDir string, outputDir string) (absInputDir string, absOutputDir string, rel string, err error) {
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
