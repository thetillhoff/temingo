package fileIO

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

func (fileList *FileList) GenerateWithIgnoreLines(ignoreLines []string, verbose bool) error {
	var (
		err       error
		filePaths []string
		ignore    *gitignore.GitIgnore
	)

	if _, err = os.Stat(fileList.Path); os.IsNotExist(err) {
		return errors.New("folder doesn't exist '" + fileList.Path + "'")
	}

	ignore = gitignore.CompileIgnoreLines(ignoreLines...)

	err = filepath.Walk(fileList.Path,
		func(relativeFilePath string, info fs.FileInfo, err error) error { // File-tree traversal function (called for each file/folder)
			if err != nil {
				return err
			}

			// As we are only checking the inputDir anyway, we can get rid of its prefix
			relativeFilePath = strings.TrimPrefix(relativeFilePath, fileList.Path)

			if relativeFilePath == "" { // do nothing if the filepath is empty (can occur for "$inputDir", after the inputDir prefix was trimmed)
				return nil
			} else if ignore != nil && ignore.MatchesPath(relativeFilePath) { // Check if ignore is set, then if it matches the current path
				// Path excluded by ignore
				if verbose {
					log.Println("Ignored because of exclusion")
				}
				if info.IsDir() { // If the current path points to a folder
					return filepath.SkipDir // Don't dive deeper in ignored folders
				}
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
