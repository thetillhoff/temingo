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
//go:embed all:testProject
var embeddedTestFilesWithPrefix embed.FS

func writeTestProjectFiles() error {
	var (
		err              error
		testProjectFiles map[string][]byte = map[string][]byte{}
	)

	err = fs.WalkDir(embeddedTestFilesWithPrefix, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			testProjectFiles[strings.TrimPrefix(path, "testProject/")], err = embeddedTestFilesWithPrefix.ReadFile(path)
			if err != nil {
				return err
			}
		}
		return err
	})

	if err != nil {
		return err
	}

	for path, content := range testProjectFiles { // For each file of the testProject (but without the path prefix)
		err = writeFile(path, content) // Write the file to disk
		if err != nil {
			return err
		}
		log.Println("File created:", path)
	}

	return nil
}
