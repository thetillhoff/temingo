package temingo

import (
	"errors"
	"log"
	"os"

	"github.com/thetillhoff/temingo/pkg/fileIO"
)

// Initializes the current directory with an example project
// It checks beforehand if the files / folders that will be written already exist and not do a thing if they do.
func (engine *Engine) InitProject(projectType string) error {
	var (
		err   error
		files map[string][]byte
	)

	if _, err := os.Stat(engine.InputDir); !os.IsNotExist(err) { // Check if the inputDir already exists
		return errors.New("the folder '" + engine.InputDir + "' already exists") // Fail if the inputdir already exists
	}

	if _, err := os.Stat(engine.TemingoignorePath); !os.IsNotExist(err) { // Check if the temingoignore already exists
		return errors.New("the file '" + engine.TemingoignorePath + "' already exists") // Fail if the temingoignore already exists
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
		if engine.Verbose {
			log.Println("File created:", path)
		}
	}
	log.Println(projectType, "project initialized.")

	return nil
}
