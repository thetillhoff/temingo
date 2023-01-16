package temingo

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// Returns the paths of all files within the inputDir (recursive traversal)
// Will remove the ignored paths before returning the list (.temingoignore)
func retrieveFilePaths(inputDir string, temingoignorePath string) ([]string, error) {
	var (
		err       error
		filePaths []string
		ignore    *gitignore.GitIgnore
	)

	if _, err = os.Stat(inputDir); os.IsNotExist(err) {
		return filePaths, errors.New("inputDir '" + inputDir + "' doesn't exist.")
	}

	if _, err = os.Stat(temingoignorePath); os.IsNotExist(err) {
		// No temingoignore
	} else if err != nil {
		// temingoignore exists, but can't be accessed
		return filePaths, err
	} else {
		// temingoignore exists and can be read
		ignore, err = gitignore.CompileIgnoreFile(temingoignorePath)
		if err != nil {
			return filePaths, err
		}
	}

	err = filepath.Walk(inputDir,
		func(relativeFilePath string, info fs.FileInfo, err error) error { // File-tree traversal function (called for each file/folder)
			if err != nil {
				return err
			}

			// As we are only checking the inputDir anyway, we can get rid of its prefix
			relativeFilePath = strings.TrimPrefix(relativeFilePath, inputDir)

			if relativeFilePath == "" { // do nothing if the filepath is empty (can occur for "$inputDir", after the inputDir prefix was trimmed)
				return nil
			} else if ignore != nil && ignore.MatchesPath(relativeFilePath) { // Check if temingoignore was found and parsed earlier, then if it matches the current path
				// Path excluded by temingoignore
				log.Println("Ignored by temingoignore: '" + relativeFilePath + "'.")
				if info.IsDir() { // If the current path points to a folder
					return filepath.SkipDir // Don't dive deeper in ignored folders
				}
			} else if info.IsDir() { // Let's keep the folders in the list, so it's easier to copy them with the correct permissions
				// Not a file, but a folder. Therefore no need to add it to the filelist.
				log.Println("Ignored because of being a folder: '" + relativeFilePath + "'.")
			} else {
				// Valid filepath
				filePaths = append(filePaths, relativeFilePath) // Add filepath to list

				log.Println("Found file: " + relativeFilePath)
			}
			return nil
		})

	return filePaths, err // Important to return err as well, as it was set during the file-tree traversal
}
