package temingo

import (
	"os"
	"path"
	"strings"

	"github.com/imdario/mergo"
)

func getMetaForDir(dirPath string, inputDir string) (interface{}, error) {
	var (
		err error

		currentFolder string
		content       []byte
		tempValues    interface{}
		values        interface{}
	)

	for _, folder := range strings.Split(path.Dir(path.Join(inputDir, dirPath)), "/") {
		currentFolder = path.Join(currentFolder, folder)
		if _, err = os.Stat(path.Join(currentFolder, "meta.yaml")); !os.IsNotExist(err) {
			content, err = readFile(path.Join(currentFolder, "meta.yaml"))
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
