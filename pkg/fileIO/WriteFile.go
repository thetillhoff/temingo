package fileIO

import (
	"os"
	"path"
)

// Creates a file at filePath and writes the content to it
// Will create all necessary parent folders
func WriteFile(filePath string, content []byte) error {
	var (
		err error
	)

	dirPath := path.Dir(filePath)

	err = os.MkdirAll(dirPath, os.ModePerm) // Create containing folder
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, content, os.ModePerm) // Create file with contents
	if err != nil {
		return err
	}

	return nil
}
