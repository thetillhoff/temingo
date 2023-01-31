package fileIO

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Returns the paths of all files within the inputDir (recursive traversal)
// Will remove the ignored paths before returning the list - does not fail if this doesn't exist!
func (fileList FileList) Generate(verbose bool) error {
	var (
		err       error
		filePaths []string
	)

	if _, err = os.Stat(fileList.Path); os.IsNotExist(err) {
		return errors.New("folder doesn't exist '" + fileList.Path + "'")
	}

	err = filepath.Walk(fileList.Path,
		func(relativeFilePath string, info fs.FileInfo, err error) error { // File-tree traversal function (called for each file/folder)
			if err != nil {
				return err
			}

			// As we are only checking the inputDir anyway, we can get rid of its prefix
			relativeFilePath = strings.TrimPrefix(relativeFilePath, fileList.Path)

			if relativeFilePath == "" { // do nothing if the filepath is empty (can occur for "$inputDir", after the inputDir prefix was trimmed)
				return nil
			} else if info.IsDir() { // Let's keep the folders in the list, so it's easier to copy them with the correct permissions // TODO fix comment
				// Not a file, but a folder. Therefore no need to add it to the filelist.
				if verbose {
					log.Println("Ignored because of being a folder: '" + relativeFilePath + "'.")
				}
			} else {
				// Valid filepath
				filePaths = append(filePaths, relativeFilePath) // Add filepath to list

				if verbose {
					log.Println("Found file: " + relativeFilePath)
				}
			}
			return nil
		})

	if err != nil {
		return err
	}

	fileList.Files = filePaths

	return nil
}
