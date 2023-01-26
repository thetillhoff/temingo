package temingo

import (
	"embed"
	"errors"
	"io/fs"
	"log"
	"path"
	"strings"
)

var (
	// While this variable contains all the files from the example projects, it has the prefix `InitFiles/` for each of the paths.
	// Do not remove the following - it configures the embedding!
	//
	//go:embed all:InitFiles
	embeddedExampleProjectFilesWithPrefix embed.FS
)

// Meant to initialize the current folder by creating some initial template files - depending on the chosen projectType (options can be retrieved by calling ProjectTypes())
// Will not write anything to disk, but returns the files as map[path]content
func (engine *Engine) getExampleProjectFiles(projectType string) (map[string][]byte, error) {
	var (
		err                 error
		exampleProjectFiles map[string][]byte = map[string][]byte{}
		treepath            string
		modifiedTreepath    string
		content             []byte
	)

	// Check if passed projectType (passed as string) is valid
	contains := false
	for _, validProjectType := range ProjectTypes() {
		if projectType == validProjectType {
			contains = true
		}
	}
	if !contains {
		return exampleProjectFiles, errors.New("not a valid project type")
	}

	if engine.Verbose {
		log.Println("Loading files from", "InitFiles/"+projectType)
	}

	err = fs.WalkDir(embeddedExampleProjectFilesWithPrefix, "InitFiles/"+projectType, func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			exampleProjectFiles[strings.TrimPrefix(filepath, "InitFiles/"+projectType+"/")], err = embeddedExampleProjectFilesWithPrefix.ReadFile(filepath)
			if err != nil {
				return err
			}
		}
		return err
	})

	if err != nil {
		return exampleProjectFiles, err
	}

	for treepath, content = range exampleProjectFiles { // For each file of the exampleProject (but without the path prefix)
		modifiedTreepath = treepath
		if strings.HasPrefix(modifiedTreepath, "src/") {
			modifiedTreepath = strings.TrimPrefix(modifiedTreepath, "src/")
			modifiedTreepath = path.Join(engine.InputDir, modifiedTreepath)
		} else if modifiedTreepath == ".temingoignore" { // No need to check if the value is actually different - simply overriding it is faster (should only be one file max anyway)
			modifiedTreepath = engine.TemingoignorePath
		} else if strings.Contains(modifiedTreepath, defaultTemplateExtension) {
			modifiedTreepath = strings.ReplaceAll(modifiedTreepath, defaultTemplateExtension, engine.TemplateExtension)
		} else if strings.Contains(modifiedTreepath, defaultMetaTemplateExtension) {
			modifiedTreepath = strings.ReplaceAll(modifiedTreepath, defaultMetaTemplateExtension, engine.MetaTemplateExtension)
		} else if strings.Contains(modifiedTreepath, defaultComponentExtension) {
			modifiedTreepath = strings.ReplaceAll(modifiedTreepath, defaultComponentExtension, engine.ComponentExtension)
		}

		if engine.Verbose {
			log.Println("Will write embedded file", treepath, "to", modifiedTreepath)
		}

		delete(exampleProjectFiles, treepath) // needs to happen before the creation of the new key-entry, else it's somehow immediately deleted
		exampleProjectFiles[modifiedTreepath] = content
	}

	return exampleProjectFiles, nil
}
