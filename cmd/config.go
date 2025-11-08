package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// loadConfig reads configuration from a YAML file in the current working directory
// It supports reading from a specific file path or defaults to .temingo.yaml in the current directory
func loadConfig(cfgFile string) (map[string]interface{}, error) {
	var configPath string

	if cfgFile != "" {
		configPath = cfgFile
	} else {
		// Default to .temingo.yaml in the current working directory
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		configPath = filepath.Join(wd, ".temingo.yaml")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, return empty config
		return make(map[string]interface{}), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Using config file:", configPath)
	return config, nil
}

// applyConfigToFlags applies config values to flag variables
// It only sets values that are not already set via command line flags (i.e., still at default values)
func applyConfigToFlags(config map[string]interface{},
	inputDirFlag, outputDirFlag, temingoignoreFlag *string,
	templateExtensionFlag, metaTemplateExtensionFlag, partialExtensionFlag *string,
	metaFilenameFlag, markdownFilenameFlag *string,
	valueFlags, valuesFileFlags *[]string,
	verboseFlag, dryRunFlag, noDeleteOutputDirFlag *bool) {
	// Helper function to get string value from config
	getString := func(key string) string {
		if val, ok := config[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
		return ""
	}

	// Helper function to get bool value from config
	getBool := func(key string) bool {
		if val, ok := config[key]; ok {
			if b, ok := val.(bool); ok {
				return b
			}
		}
		return false
	}

	// Helper function to get string slice from config
	getStringSlice := func(key string) []string {
		if val, ok := config[key]; ok {
			if slice, ok := val.([]interface{}); ok {
				result := make([]string, 0, len(slice))
				for _, v := range slice {
					if str, ok := v.(string); ok {
						result = append(result, str)
					}
				}
				return result
			}
		}
		return nil
	}

	// Apply config values only if flags are still at default values
	if *inputDirFlag == "src/" {
		if val := getString("inputDir"); val != "" {
			*inputDirFlag = val
		}
	}
	if *outputDirFlag == "output/" {
		if val := getString("outputDir"); val != "" {
			*outputDirFlag = val
		}
	}
	if *temingoignoreFlag == ".temingoignore" {
		if val := getString("temingoignore"); val != "" {
			*temingoignoreFlag = val
		}
	}
	if *templateExtensionFlag == ".template" {
		if val := getString("templateExtension"); val != "" {
			*templateExtensionFlag = val
		}
	}
	if *metaTemplateExtensionFlag == ".metatemplate" {
		if val := getString("metaTemplateExtension"); val != "" {
			*metaTemplateExtensionFlag = val
		}
	}
	if *partialExtensionFlag == ".partial" {
		if val := getString("partialExtension"); val != "" {
			*partialExtensionFlag = val
		}
	}
	if *metaFilenameFlag == "meta.yaml" {
		if val := getString("metaFilename"); val != "" {
			*metaFilenameFlag = val
		}
	}
	if *markdownFilenameFlag == "content.md" {
		if val := getString("markdownFilename"); val != "" {
			*markdownFilenameFlag = val
		}
	}
	if !*verboseFlag {
		*verboseFlag = getBool("verbose")
	}
	if !*dryRunFlag {
		*dryRunFlag = getBool("dryRun")
	}
	if !*noDeleteOutputDirFlag {
		*noDeleteOutputDirFlag = getBool("noDeleteOutputDir")
	}

	// Handle string slices - use config values if CLI values are empty
	if len(*valueFlags) == 0 {
		if configValues := getStringSlice("value"); len(configValues) > 0 {
			*valueFlags = configValues
		}
	}
	if len(*valuesFileFlags) == 0 {
		if configValues := getStringSlice("valuesfile"); len(configValues) > 0 {
			*valuesFileFlags = configValues
		}
	}
}
