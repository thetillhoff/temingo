package temingo

import (
	"os"
)

// Returns the contents of the file at the provided path
func readFile(filePath string) ([]byte, error) {
	var (
		err         error
		fileContent []byte
	)

	fileContent, err = os.ReadFile(filePath)

	return fileContent, err
}
