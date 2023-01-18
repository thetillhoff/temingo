package temingo

import (
	"embed"
	"io/fs"
	"log"
	"path"
	"strings"
)

// While this variable contains all the files from the example project, it has the prefix `exampleProject/` for each of the paths.
// Do not remove the following - it configures the embedding!
//
//go:embed all:exampleProject
var embeddedExampleFilesWithPrefix embed.FS

func writeExampleProjectFiles(inputDir string) error {
	var (
		err                 error
		exampleProjectFiles map[string][]byte = map[string][]byte{}
	)

	err = fs.WalkDir(embeddedExampleFilesWithPrefix, ".", func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			exampleProjectFiles[strings.TrimPrefix(filepath, "exampleProject/")], err = embeddedExampleFilesWithPrefix.ReadFile(filepath)
			if err != nil {
				return err
			}
		}
		return err
	})

	if err != nil {
		return err
	}

	for treepath, content := range exampleProjectFiles { // For each file of the exampleProject (but without the path prefix)
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
