package temingo

import (
	"log"
)

func Init() error {
	err := WriteExampleProjectFiles()
	if err != nil {
		return err
	}

	log.Println("Project initialized.")

	return nil
}
