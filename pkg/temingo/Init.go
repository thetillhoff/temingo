package temingo

import (
	"errors"
	"log"
	"os"
)

func Init(inputDirFlag string, temingoignorePathFlag string, templateExtensionFlag string, metaTemplateExtensionFlag string, componentExtensionFlag string, verboseFlag bool, projectType string) error {
	var (
		err   error
		files map[string][]byte
	)

	// Set flags globally so they don't have to be passed around all the time
	inputDir = inputDirFlag
	temingoignorePath = temingoignorePathFlag
	templateExtension = templateExtensionFlag
	metaTemplateExtension = metaTemplateExtensionFlag
	componentExtension = componentExtensionFlag
	verbose = verboseFlag

	if _, err := os.Stat(inputDir); !os.IsNotExist(err) { // Check if the inputDir already exists
		return errors.New("the folder '" + inputDir + "' already exists") // Fail if the inputdir already exists
	}

	files, err = writeExampleProjectFiles(projectType)
	if err != nil {
		return err
	}

	for path, content := range files {
		err = writeFile(path, content) // Write the file to disk
		if err != nil {
			return err
		}
		if verbose {
			log.Println("File created:", path)
		}
	}

	log.Println("Project initialized.")

	return nil
}
