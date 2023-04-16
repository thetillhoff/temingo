package temingo

import (
	"log"
	"path"

	"github.com/imdario/mergo"
	"github.com/thetillhoff/fileIO"
	"gopkg.in/yaml.v3"
)

// Reads the meta data for the specified templatePath
// Returns the metadata from the treePath (meta yamls of the template-dir and the direct children-dirs),
// the childMetadata in map[folderName]metadata format
func (engine *Engine) getMetaForTemplatePath(metaTemplatePaths fileIO.FileList, templatePath string) (map[string]interface{}, map[string]interface{}, error) {
	var (
		err       error
		meta      map[string]interface{} = map[string]interface{}{}
		childMeta map[string]interface{} = map[string]interface{}{}

		content       []byte
		parsedContent interface{}
		folderName    string
	)

	for _, metaFilePath := range metaTemplatePaths.FilterByTreePath(templatePath).Files { // For each meta yaml in dirTree for templatePath
		if engine.Verbose {
			log.Println("Reading metadata from", metaFilePath)
		}

		content, err = fileIO.ReadFile(path.Join(engine.InputDir, metaFilePath)) // Read file contents
		if err != nil {
			return nil, nil, err
		}
		err = yaml.Unmarshal(content, &parsedContent) // Store yaml into map
		if err != nil {
			return nil, nil, err
		}

		err := mergo.Merge(&meta, parsedContent, mergo.WithOverride) // Merge while overriding existing values
		if err != nil {
			return nil, nil, err
		}
	}

	for _, childMetaFilePath := range metaTemplatePaths.FilterByLevelAtFolderPath(path.Dir(templatePath), 1).Files { // For each direct child meta yaml
		if engine.Verbose {
			log.Println("Reading child-metadata from", childMetaFilePath)
		}

		content, err = fileIO.ReadFile(path.Join(engine.InputDir, childMetaFilePath)) // Read file contents
		if err != nil {
			return nil, nil, err
		}
		err = yaml.Unmarshal(content, &parsedContent) // Store yaml into map
		if err != nil {
			return nil, nil, err
		}

		folderName = path.Base(path.Dir(childMetaFilePath)) // Get the name of the last folder

		childMeta[folderName] = parsedContent
	}

	return meta, childMeta, nil
}
