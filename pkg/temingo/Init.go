package temingo

import (
	"errors"
	"log"
	"os"
)

func Init(inputDir string, temingoignore string) error {
	var (
		err error
	)

	if _, err := os.Stat(inputDir); !os.IsNotExist(err) { // Check if the inputDir already exists
		return errors.New("the folder '" + inputDir + "' already exists") // Fail if the inputdir already exists
	}

	// TODO change below to writeExampleProjectfiles instead of writeTestProjectFiles
	// err = writeExampleProjectFiles(inputDir)
	err = writeTestProjectFiles(inputDir, temingoignore)
	if err != nil {
		return err
	}

	log.Println("Project initialized.")

	return nil
}
