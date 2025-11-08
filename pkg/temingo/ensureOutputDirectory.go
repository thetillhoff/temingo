package temingo

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// ensureOutputDirectory checks that outputDir exists and is a directory, or creates it if it doesn't exist.
// This is a simpler validation function for use cases like Init that don't need input/output comparison.
func ensureOutputDirectory(outputDir string, logger *slog.Logger) error {
	// Initialize logger if not provided (use default logger as fallback)
	if logger == nil {
		logger = slog.Default()
	}

	// Handle trailing separator - if path ends with separator, try without it first
	outputDirToCheck := strings.TrimSuffix(outputDir, string(filepath.Separator))
	if outputDirToCheck == "" {
		outputDirToCheck = outputDir
	}

	// Normalize the path to ensure consistent handling
	outputDirToCheck = filepath.Clean(outputDirToCheck)
	info, err := os.Stat(outputDirToCheck)
	if err != nil {
		if os.IsNotExist(err) {
			// Create output directory with default permissions (0755)
			// Use the cleaned path without separator for MkdirAll
			err = os.MkdirAll(outputDirToCheck, 0755)
			if err != nil {
				return fmt.Errorf("error creating output directory %s: %w", outputDir, err)
			}

			// Use Chmod to ensure exact permissions (MkdirAll may be affected by umask)
			err = os.Chmod(outputDirToCheck, 0755)
			if err != nil {
				return fmt.Errorf("error setting output directory permissions %s: %w", outputDir, err)
			}
			logger.Info("Created output directory", "path", outputDir)
		} else {
			return fmt.Errorf("error accessing output directory %s: %w", outputDir, err)
		}
	} else if !info.IsDir() {
		return fmt.Errorf("output directory is not a directory: %s", outputDir)
	}
	return nil
}
