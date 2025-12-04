package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
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
// Precedence: CLI/env flags > config file > defaults
// Config values are applied first, then CLI/env values override if they were explicitly set
func applyConfigToFlags(cmd *cli.Command, config map[string]interface{},
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

	// Helper function to check if a flag was explicitly set
	isFlagSet := func(name string) bool {
		for _, flag := range cmd.Flags {
			for _, flagName := range flag.Names() {
				if flagName == name {
					return flag.IsSet()
				}
			}
		}
		return false
	}

	// Helper function to apply config and CLI/env values for a string flag
	applyStringFlag := func(flagName, configKey string, target *string) {
		if val := getString(configKey); val != "" {
			*target = val
		}
		if isFlagSet(flagName) {
			*target = cmd.String(flagName)
		}
	}

	// Helper function to apply config and CLI/env values for a bool flag
	applyBoolFlag := func(flagName, configKey string, target *bool) {
		*target = getBool(configKey)
		if isFlagSet(flagName) {
			*target = cmd.Bool(flagName)
		}
	}

	// Helper function to apply config and CLI/env values for a string slice flag
	applyStringSliceFlag := func(flagName, configKey string, target *[]string) {
		if configValues := getStringSlice(configKey); len(configValues) > 0 {
			*target = configValues
		}
		if isFlagSet(flagName) {
			*target = cmd.StringSlice(flagName)
		}
	}

	// Apply config and CLI/env values for all flags
	applyStringFlag("inputDir", "inputDir", inputDirFlag)
	applyStringFlag("outputDir", "outputDir", outputDirFlag)
	applyStringFlag("temingoignore", "temingoignore", temingoignoreFlag)
	applyStringFlag("templateExtension", "templateExtension", templateExtensionFlag)
	applyStringFlag("metaTemplateExtension", "metaTemplateExtension", metaTemplateExtensionFlag)
	applyStringFlag("partialExtension", "partialExtension", partialExtensionFlag)
	applyStringFlag("metaFilename", "metaFilename", metaFilenameFlag)
	applyStringFlag("markdownFilename", "markdownFilename", markdownFilenameFlag)
	applyBoolFlag("verbose", "verbose", verboseFlag)
	applyBoolFlag("dry-run", "dryRun", dryRunFlag)
	applyBoolFlag("noDeleteOutputDir", "noDeleteOutputDir", noDeleteOutputDirFlag)
	applyStringSliceFlag("value", "value", valueFlags)
	applyStringSliceFlag("valuesfile", "valuesfile", valuesFileFlags)
}
