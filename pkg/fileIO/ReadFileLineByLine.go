package fileIO

import (
	"os"
	"strings"
)

// Returns the contents of the file at the provided path
func ReadFileLineByLine(filePath string) ([]string, error) {
	var (
		err          error
		fileContents []byte
		lines        []string
	)

	fileContents, err = os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines = strings.Split(string(fileContents), "\n")

	return lines, err
}
