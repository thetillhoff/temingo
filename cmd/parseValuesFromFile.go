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

// parseValuesFromFiles reads multiple YAML files and merges them into a single map
// Files are merged in order, with later files overriding earlier ones
func parseValuesFromFiles(filePaths []string) (map[string]string, error) {
	mergedValues := make(map[string]string)

	for _, filePath := range filePaths {
		fileValues, err := parseValuesFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error parsing values file %s: %w", filePath, err)
		}

		// Merge with existing values (later files override earlier ones)
		for key, value := range fileValues {
			mergedValues[key] = value
		}
	}

	return mergedValues, nil
}
