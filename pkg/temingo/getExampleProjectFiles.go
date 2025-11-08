package temingo

import (
	"embed"
	"errors"
	"io/fs"
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
	logger := engine.Logger

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

	logger.Debug("Loading files from embedded project", "path", "InitFiles/"+projectType)

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
		if strings.HasPrefix(modifiedTreepath, "src/") { // Needs to be done for all the files except the ones in the root dir like the .temingoignore
			modifiedTreepath = strings.TrimPrefix(modifiedTreepath, "src/")
			modifiedTreepath = path.Join(engine.InputDir, modifiedTreepath)
		} else if modifiedTreepath == ".temingoignore" { // No need to check if the value is actually different - simply overriding it is faster (should only be one file max anyway)
			modifiedTreepath = engine.TemingoignorePath
		}

		if strings.Contains(modifiedTreepath, defaultTemplateExtension) {
			modifiedTreepath = strings.ReplaceAll(modifiedTreepath, defaultTemplateExtension, engine.TemplateExtension)
		} else if strings.Contains(modifiedTreepath, defaultMetaTemplateExtension) {
			modifiedTreepath = strings.ReplaceAll(modifiedTreepath, defaultMetaTemplateExtension, engine.MetaTemplateExtension)
		} else if strings.Contains(modifiedTreepath, defaultPartialExtension) {
			modifiedTreepath = strings.ReplaceAll(modifiedTreepath, defaultPartialExtension, engine.PartialExtension)
		} else if path.Base(modifiedTreepath) == defaultMetaFilename {
			modifiedTreepath = path.Join(path.Dir(modifiedTreepath), engine.MetaFilename)
		}

		logger.Debug("Will write embedded file", "from", treepath, "to", modifiedTreepath)

		delete(exampleProjectFiles, treepath) // needs to happen before the creation of the new key-entry, else it's somehow immediately deleted
		exampleProjectFiles[modifiedTreepath] = content
	}

	return exampleProjectFiles, nil
}
