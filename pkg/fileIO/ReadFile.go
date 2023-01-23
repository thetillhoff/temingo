package fileIO

import (
	"os"
)

// Returns the contents of the file at the provided path
func ReadFile(filePath string) ([]byte, error) {
	var (
		err         error
		fileContent []byte
	)

	fileContent, err = os.ReadFile(filePath)

	return fileContent, err
}
