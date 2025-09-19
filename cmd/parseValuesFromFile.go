package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// parseValuesFromFile reads a YAML file and returns a map of string key-value pairs
// for use in template rendering
func parseValuesFromFile(filePath string) (map[string]string, error) {
	var yamlData map[string]interface{}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return nil, err
	}

	// Convert to map[string]string
	result := make(map[string]string)
	for key, value := range yamlData {
		if strValue, ok := value.(string); ok {
			result[key] = strValue
		} else {
			// Convert non-string values to string representation
			result[key] = fmt.Sprintf("%v", value)
		}
	}

	return result, nil
}
