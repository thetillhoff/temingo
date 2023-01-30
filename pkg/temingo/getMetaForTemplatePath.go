package temingo

import (
	"log"
	"path"

	"github.com/imdario/mergo"
	"github.com/thetillhoff/temingo/pkg/fileIO"
	"gopkg.in/yaml.v3"
)

// TODO merge getMetaForPath and the related part of renderTemplate into this function
// Returns map["meta"] and map["childMeta"]

// uses fileList to determine which files to read
// reads the meta yamls of the template-dir and the direct children-dirs

func (engine *Engine) getMetaForTemplatePath(metaTemplatePaths fileIO.FileList, templatePath string) (map[string]interface{}, error) {
	var (
		err  error
		meta map[string]interface{} = map[string]interface{}{}

		content       []byte
		parsedContent interface{}
	)

	log.Println("Getting metadata for", templatePath)
	log.Println("metadata:", metaTemplatePaths.Files)
	log.Println("filtered metadatapaths:", metaTemplatePaths.FilterByTreePath(path.Dir(templatePath)).Files)

	for _, metaFilePath := range metaTemplatePaths.FilterByTreePath(path.Dir(templatePath)).Files { // For each meta.yaml in dirTree for templatePath
		if engine.Verbose {
			log.Println("Reading metadata from", metaFilePath)
		}

		content, err = fileIO.ReadFile(metaFilePath) // Read file contents
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(content, &parsedContent) // Store yaml into map
		if err != nil {
			return nil, err
		}

		err := mergo.Merge(meta["meta"], parsedContent, mergo.WithOverride) // Merge while overriding existing values
		if err != nil {
			return nil, err
		}
	}

	for _, childMetaFilePath := range metaTemplatePaths.FilterByLevelAtFolderPath(path.Dir(templatePath), 1).Files { // For each direct child meta.yaml
		if engine.Verbose {
			log.Println("Reading child-metadata from", childMetaFilePath)
		}

		// TODO
		// get folderName for map[folderName]meta
		// read filecontents
		// put contents into map["childMeta"][folderName]

		// 	content, err = fileIO.ReadFile(childMetaFilePath) // Read file contents
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	err = yaml.Unmarshal(content, &parsedContent) // Store yaml into map
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	err := mergo.Merge(meta["meta"], parsedContent, mergo.WithOverride) // Merge while overriding existing values
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }

		// for _, f := range files { // For each child-element of folder
		// 	if f.IsDir() { // Only for folders
		// 		if engine.Verbose {
		// 			log.Println("Searching child metadata for", engine.InputDir+path.Join(path.Dir(templatePath), f.Name()))
		// 		}
		// 		tempMeta, err = getMetaForPath(fileList, path.Join(path.Dir(templatePath), f.Name()), engine.Verbose)
		// 		if err != nil {
		// 			return nil, err
		// 		}
		// 		if tempMeta != nil {
		// 			childMetaForDir[f.Name()] = tempMeta
		// 		}
		// 	}
		// }

	}
	return meta, nil
}
