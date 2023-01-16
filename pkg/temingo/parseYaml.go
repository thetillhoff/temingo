package temingo

import (
	"gopkg.in/yaml.v3"
)

// Returns the object contained in the provided yaml
func parseYaml(content []byte) (map[string]interface{}, error) {
	var (
		err          error
		mappedObject map[string]interface{}
	)

	err = yaml.Unmarshal(content, &mappedObject) // store yaml into map

	return mappedObject, err
}
