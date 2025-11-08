package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ValidateDirectories checks that inputDir exists and is a directory,
// and that outputDir exists and is a directory (or creates it if it doesn't exist)
func ValidateDirectories(inputDir string, outputDir string) error {
	// Validate inputDir
	info, err := os.Stat(inputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", inputDir)
		}
		return fmt.Errorf("error accessing input directory %s: %w", inputDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("input directory is not a directory: %s", inputDir)
	}

	// Validate or create outputDir
	info, err = os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Create output directory if it doesn't exist
			err = os.MkdirAll(outputDir, 0755)
			if err != nil {
				return fmt.Errorf("error creating output directory %s: %w", outputDir, err)
			}
			log.Printf("Created output directory: %s", outputDir)
		} else {
			return fmt.Errorf("error accessing output directory %s: %w", outputDir, err)
		}
	} else if !info.IsDir() {
		return fmt.Errorf("output directory is not a directory: %s", outputDir)
	}

	// Ensure outputDir is absolute for better error messages
	absOutputDir, err := filepath.Abs(outputDir)
	if err == nil {
		absInputDir, err := filepath.Abs(inputDir)
		if err == nil {
			// Check if outputDir is inside inputDir (which would cause issues)
			rel, err := filepath.Rel(absInputDir, absOutputDir)
			if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return fmt.Errorf("output directory cannot be inside or equal to input directory: output=%s, input=%s", absOutputDir, absInputDir)
			}
		}
	}

	return nil
}

