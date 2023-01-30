package temingo

import (
	"log"

	"github.com/imdario/mergo"
	"github.com/thetillhoff/temingo/pkg/fileIO"
	"gopkg.in/yaml.v3"
)

// Reads all meta files from all directories from the root-dirPath to dirPath and merges them.
// Returns a interface{} of the result.
func getMetaForPath(fileList fileIO.FileList, dirPath string, verbose bool) (interface{}, error) {
	var (
		err error

		content    []byte
		tempValues interface{}
		values     interface{}
	)

	for _, metaFilePath := range fileList.FilterByTreePath(dirPath).FilterByFileName(defaultMetaFileName).Files {
		if verbose {
			log.Println("Reading metadata from", metaFilePath)
		}

		content, err = fileIO.ReadFile(metaFilePath)
		if err != nil {
			return values, err
		}
		err = yaml.Unmarshal(content, &tempValues) // Store yaml into map
		if err != nil {
			return values, err
		}

		err := mergo.Merge(&values, tempValues, mergo.WithOverride) // Merge while overriding existing values
		if err != nil {
			return values, err
		}
	}

	return values, nil
}
