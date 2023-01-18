package temingo

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/imdario/mergo"
)

func getMetaForDir(startDir string, dirPath string, verbose bool) (map[string]interface{}, error) {
	var (
		err error

		currentFolder       = ""
		currentMetaLocation string
		content             []byte
		tempValues          map[string]interface{}
		values              map[string]interface{}
	)

	for _, folder := range strings.Split(path.Join(startDir, dirPath), "/") {
		currentFolder = path.Join(currentFolder, folder)
		currentMetaLocation = path.Join(currentFolder, "meta.yaml")
		if _, err = os.Stat(currentMetaLocation); !os.IsNotExist(err) {
			if verbose {
				log.Println("Reading metadata from", currentMetaLocation)
			}
			content, err = readFile(currentMetaLocation)
			if err != nil {
				return values, err
			}
			tempValues, err = parseYaml(content)
			if err != nil {
				return values, err
			}

			err := mergo.Merge(&values, tempValues, mergo.WithOverride)
			if err != nil {
				return values, err
			}
		}
	}

	return values, nil
}
