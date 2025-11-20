package temingo

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/thetillhoff/fileIO"
)

// Initializes the targetDir directory with an example project
// It checks beforehand if the files / folders that will be written already exist and not do a thing if they do.
func (engine *Engine) InitProject(projectType string, targetDir string) error {
	logger := engine.Logger

	var (
		err   error
		files map[string][]byte
	)

	// Ensure targetDir is absolute
	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		return errors.New("error converting targetDir to absolute: " + err.Error())
	}

	// Ensure output directory exists (relative to targetDir)
	outputDirPath := filepath.Join(targetDir, engine.OutputDir)
	err = ensureOutputDirectory(outputDirPath, logger)
	if err != nil {
		return err
	}

	// Get absolute path for output directory exclusion
	outputDirAbs, err := filepath.Abs(outputDirPath)
	if err != nil {
		outputDirAbs = ""
	}

	// Check for specific files/directories first (more specific errors)
	// These will be created by InitProject from embedded files, so they must not exist
	// Paths are relative to targetDir, so join them with targetDir for checking
	inputDirPath := filepath.Join(targetDir, engine.InputDir)
	if _, err := os.Stat(inputDirPath); !os.IsNotExist(err) {
		return errors.New("the folder '" + engine.InputDir + "' already exists")
	}

	temingoignorePath := filepath.Join(targetDir, engine.TemingoignorePath)
	if _, err := os.Stat(temingoignorePath); !os.IsNotExist(err) {
		return errors.New("the file '" + engine.TemingoignorePath + "' already exists")
	}
	if err != nil {
		outputDirAbs = ""
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return errors.New("error reading target directory: " + err.Error())
	}

	// Filter out only the output directory from the entries
	nonEmpty := false
	for _, entry := range entries {
		entryPath := filepath.Join(targetDir, entry.Name())
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

	// Normalize all paths to be relative to targetDir
	// If a path is absolute, make it relative to targetDir
	// If a path is relative, keep it as-is
	normalizedFiles := make(map[string][]byte)
	for path, content := range files {
		var relPath string
		if filepath.IsAbs(path) {
			// Path is absolute - make it relative to targetDir
			relPath, err = filepath.Rel(targetDir, path)
			if err != nil {
				return errors.New("error making path relative to targetDir: " + err.Error())
			}
		} else {
			// Path is already relative, use it as-is
			relPath = path
		}
		normalizedFiles[relPath] = content
	}
	files = normalizedFiles

	// Write files - join relative paths with targetDir to get absolute paths
	for relPath, content := range files {
		absPath := filepath.Join(targetDir, relPath)
		err = fileIO.WriteFile(absPath, content) // Write the file to disk
		if err != nil {
			return err
		}
		logger.Debug("File created", "path", absPath)
	}
	logger.Info("Project initialized", "type", projectType)

	return nil
}
