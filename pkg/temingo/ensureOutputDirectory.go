package temingo

import (
	"fmt"
	"log"
	"os"
)

// ensureOutputDirectory checks that outputDir exists and is a directory, or creates it if it doesn't exist.
// This is a simpler validation function for use cases like Init that don't need input/output comparison.
func ensureOutputDirectory(outputDir string) error {
	info, err := os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Create output directory with default permissions (0755)
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
	return nil
}

