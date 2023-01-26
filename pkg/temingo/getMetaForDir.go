package temingo

import (
	"log"

	"github.com/imdario/mergo"
	"github.com/thetillhoff/temingo/pkg/fileIO"
	"gopkg.in/yaml.v3"
)

func getMetaForDir(fileList fileIO.FileList, startDir string, dirPath string, verbose bool) (map[string]interface{}, error) {
	var (
		err error

		content    []byte
		tempValues map[string]interface{}
		values     map[string]interface{}
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
