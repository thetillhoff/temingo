package temingo

import (
	"gopkg.in/yaml.v2"
)

// Returns the object contained in the provided yaml
func parseYaml(content string) (map[string]interface{}, error) {
	var (
		err          error
		mappedObject map[string]interface{}
	)

	err = yaml.Unmarshal([]byte(content), &mappedObject) // store yaml into map

	return mappedObject, err
}
