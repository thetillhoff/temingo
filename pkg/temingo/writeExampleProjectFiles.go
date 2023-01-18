package temingo

import (
	"embed"
	"io/fs"
	"log"
	"path"
	"strings"
)

var (
	// While this variable contains all the files from the example project, it has the prefix `exampleProject/` for each of the paths.
	// Do not remove the following - it configures the embedding!
	//
	//go:embed all:exampleProject
	embeddedExampleFilesWithPrefix embed.FS
)

func writeExampleProjectFiles(inputDir string, temingoignore string, templateExtension string, metaTemplateExtension string, componentExtension string, verbose bool) error {
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
		} else if treepath == ".temingoignore" { // No need to check if the value is actually different - simply overriding it is faster
			treepath = temingoignore
		} else if strings.Contains(treepath, defaultTemplateExtension) {
			treepath = strings.ReplaceAll(treepath, defaultTemplateExtension, templateExtension)
		} else if strings.Contains(treepath, defaultMetaTemplateExtension) {
			treepath = strings.ReplaceAll(treepath, defaultMetaTemplateExtension, metaTemplateExtension)
		} else if strings.Contains(treepath, defaultComponentExtension) {
			treepath = strings.ReplaceAll(treepath, defaultComponentExtension, componentExtension)
		}

		err = writeFile(treepath, content) // Write the file to disk
		if err != nil {
			return err
		}
		if verbose {
			log.Println("File created:", treepath)
		}
	}

	return nil
}
