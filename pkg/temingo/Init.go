package temingo

import (
	"log"
)

func Init() error {
	var (
		err error
	)
	err = writeExampleProjectFiles()
	// err = writeTestProjectFiles()
	if err != nil {
		return err
	}

	log.Println("Project initialized.")

	return nil
}
