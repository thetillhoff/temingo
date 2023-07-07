package temingo

import (
	"log"
	"path"

	"github.com/thetillhoff/fileIO"
	"github.com/thetillhoff/temingo/pkg/mergeYaml"
	"gopkg.in/yaml.v3"
)

// Reads the meta data for the specified templatePath
// Returns the metadata from the treePath (meta yamls of the template-dir and the direct children-dirs),
// the childMetadata in map[folderName]metadata format
func (engine *Engine) getMetaForTemplatePath(metaTemplatePaths fileIO.FileList, templatePath string) (interface{}, map[string]interface{}, error) {
	var (
		err       error
		meta      interface{}                                       // meta object for templatepath -> {meta}
		childMeta map[string]interface{} = map[string]interface{}{} // meta object for each childPath -> childMeta[childname]{meta}

		metaContent   []byte
		parsedContent interface{}

		folderName string
	)

	for _, metaFilePath := range metaTemplatePaths.FilterByTreePath(templatePath).Files { // For each meta yaml in dirTree for templatePath (top-down)
		if engine.Verbose {
			log.Println("Reading metadata from", metaFilePath)
		}

		metaContent, err = fileIO.ReadFile(path.Join(engine.InputDir, metaFilePath)) // Read file contents
		if err != nil {
			return nil, nil, err
		}
		err = yaml.Unmarshal(metaContent, &parsedContent) // Store yaml into map
		if err != nil {
			return nil, nil, err
		}

		meta = mergeYaml.Merge(parsedContent, meta, true)
	}

	for _, childMetaFilePath := range metaTemplatePaths.FilterByLevelAtFolderPath(path.Dir(templatePath), 1).Files { // For each direct child meta yaml
		if engine.Verbose {
			log.Println("Reading child-metadata from", childMetaFilePath)
		}

		metaContent, err = fileIO.ReadFile(path.Join(engine.InputDir, childMetaFilePath)) // Read file contents
		if err != nil {
			return nil, nil, err
		}
		err = yaml.Unmarshal(metaContent, &parsedContent) // Store yaml into map
		if err != nil {
			return nil, nil, err
		}

		folderName = path.Base(path.Dir(childMetaFilePath)) // Get the name of the last folder

		childMeta[folderName] = mergeYaml.Merge(parsedContent, meta, true) // Store parent+child meta into childMeta objects per child-folder
	}

	return meta, childMeta, nil
}
