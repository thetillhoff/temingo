package temingo

import (
	"embed"
	"io/fs"
	"log"
	"strings"
)

// While this variable contains all the files from the example project, it has the prefix `exampleProject/` for each of the paths.
// Do not remove the following - it configures the embedding!
//
//go:embed exampleProject
var embeddedExampleFilesWithPrefix embed.FS

func writeExampleProjectFiles() error {
	var (
		err                 error
		exampleProjectFiles map[string][]byte = map[string][]byte{}
	)

	err = fs.WalkDir(embeddedExampleFilesWithPrefix, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			exampleProjectFiles[strings.TrimPrefix(path, "exampleProject/")], err = embeddedExampleFilesWithPrefix.ReadFile(path)
			if err != nil {
				return err
			}
		}
		return err
	})

	if err != nil {
		return err
	}

	for path, content := range exampleProjectFiles { // For each file of the exampleProject (but without the path prefix)
		err = writeFile(path, content) // Write the file to disk
		if err != nil {
			return err
		}
		log.Println("File created:", path)
	}

	return nil
}
