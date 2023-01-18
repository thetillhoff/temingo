package temingo

import (
	"embed"
	"io/fs"
	"log"
	"path"
	"strings"
)

// While this variable contains all the files from the test project, it has the prefix `testProject/` for each of the paths.
// Do not remove the following - it configures the embedding!
//
//go:embed all:testProject
var embeddedTestFilesWithPrefix embed.FS

func writeTestProjectFiles(inputDir string) error {
	var (
		err              error
		testProjectFiles map[string][]byte = map[string][]byte{}
	)

	err = fs.WalkDir(embeddedTestFilesWithPrefix, ".", func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			testProjectFiles[strings.TrimPrefix(filepath, "testProject/")], err = embeddedTestFilesWithPrefix.ReadFile(filepath)
			if err != nil {
				return err
			}
		}
		return err
	})

	if err != nil {
		return err
	}

	for treepath, content := range testProjectFiles { // For each file of the testProject (but without the path prefix)
		if strings.HasPrefix(treepath, "src/") {
			treepath = strings.TrimPrefix(treepath, "src/")
			treepath = path.Join(inputDir, treepath)
		}
		err = writeFile(treepath, content) // Write the file to disk
		if err != nil {
			return err
		}
		log.Println("File created:", treepath)
	}

	return nil
}
