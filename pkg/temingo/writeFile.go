package temingo

// This is automatically tested if `WriteExampleProjectFiles` is tested.

import (
	"os"
	"path"
)

func writeFile(filePath string, content []byte) error {
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
