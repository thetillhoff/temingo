package temingo

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/thetillhoff/fileIO"
)

// Initializes the current directory with an example project
// It checks beforehand if the files / folders that will be written already exist and not do a thing if they do.
func (engine *Engine) InitProject(projectType string) error {
	logger := engine.Logger

	var (
		err   error
		files map[string][]byte
	)

	// Ensure output directory exists
	err = ensureOutputDirectory(engine.OutputDir, logger)
	if err != nil {
		return err
	}

	// Check for specific files/directories first (more specific errors)
	// These will be created by InitProject from embedded files, so they must not exist
	if _, err := os.Stat(engine.InputDir); !os.IsNotExist(err) {
		return errors.New("the folder '" + engine.InputDir + "' already exists")
	}

	if _, err := os.Stat(engine.TemingoignorePath); !os.IsNotExist(err) {
		return errors.New("the file '" + engine.TemingoignorePath + "' already exists")
	}

	// Check if the current working directory is not empty (excluding output directory)
	// InputDir and TemingoignorePath don't need to be excluded since we already verified they don't exist
	cwd, err := os.Getwd()
	if err != nil {
		return errors.New("error getting current working directory: " + err.Error())
	}

	// Get absolute path for output directory exclusion
	outputDirToCheck := filepath.Clean(engine.OutputDir)
	if outputDirToCheck == "" || outputDirToCheck == "." {
		outputDirToCheck = engine.OutputDir
	}
	outputDirAbs, err := filepath.Abs(outputDirToCheck)
	if err != nil {
		outputDirAbs = ""
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		return errors.New("error reading current directory: " + err.Error())
	}

	// Filter out only the output directory from the entries
	nonEmpty := false
	for _, entry := range entries {
		entryPath := filepath.Join(cwd, entry.Name())
		entryAbs, err := filepath.Abs(entryPath)
		if err != nil {
			continue
		}
		// Skip the output directory if it exists (it's allowed to exist)
		if outputDirAbs != "" && entryAbs == outputDirAbs {
			continue
		}
		nonEmpty = true
		break
	}

	if nonEmpty {
		return errors.New("the directory is not empty")
	}

	files, err = engine.getExampleProjectFiles(projectType)
	if err != nil {
		return err
	}

	for path, content := range files {
		err = fileIO.WriteFile(path, content) // Write the file to disk
		if err != nil {
			return err
		}
		logger.Debug("File created", "path", path)
	}
	logger.Info("Project initialized", "type", projectType)

	return nil
}
