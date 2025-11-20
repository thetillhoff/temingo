package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/thetillhoff/temingo/pkg/temingo"
	"github.com/urfave/cli/v3"
)

// initCommand represents the init command
var initCommand = &cli.Command{
	Name:      "init",
	Usage:     "Initializes the current directory with an example project",
	UsageText: "temingo init {" + strings.Join(temingo.ProjectTypes(), "|") + "}",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		args := cmd.Args()
		if args.Len() != 1 {
			return fmt.Errorf("init requires exactly one argument: %s", strings.Join(temingo.ProjectTypes(), ", "))
		}

		projectType := args.Get(0)

		// Validate project type
		validTypes := temingo.ProjectTypes()
		valid := false
		for _, t := range validTypes {
			if t == projectType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid project type. Available types are: %s", strings.Join(validTypes, ", "))
		}

		cfgFile := cmd.String("config")
		inputDirFlag := cmd.String("inputDir")
		outputDirFlag := cmd.String("outputDir")
		temingoignoreFlag := cmd.String("temingoignore")
		templateExtensionFlag := cmd.String("templateExtension")
		metaTemplateExtensionFlag := cmd.String("metaTemplateExtension")
		partialExtensionFlag := cmd.String("partialExtension")
		metaFilenameFlag := cmd.String("metaFilename")
		markdownFilenameFlag := cmd.String("markdownFilename")
		valueFlags := cmd.StringSlice("value")
		valuesFileFlags := cmd.StringSlice("valuesfile")
		verboseFlag := cmd.Bool("verbose")
		dryRunFlag := cmd.Bool("dry-run")
		noDeleteOutputDirFlag := cmd.Bool("noDeleteOutputDir")

		// Load config file if specified
		config, err := loadConfig(cfgFile)
		if err != nil {
			slog.Error("Failed to load config", "error", err)
			return err
		}

		// Apply config values to flags
		applyConfigToFlags(config, &inputDirFlag, &outputDirFlag, &temingoignoreFlag,
			&templateExtensionFlag, &metaTemplateExtensionFlag, &partialExtensionFlag,
			&metaFilenameFlag, &markdownFilenameFlag, &valueFlags, &valuesFileFlags,
			&verboseFlag, &dryRunFlag, &noDeleteOutputDirFlag)

		var (
			values = map[string]string{}
		)

		if !strings.HasSuffix(inputDirFlag, "/") {
			inputDirFlag += "/"
		}
		if !strings.HasSuffix(outputDirFlag, "/") {
			outputDirFlag += "/"
		}

		if len(valuesFileFlags) > 0 {
			// Parse values from files first (merge multiple files)
			values, err = parseValuesFromFiles(valuesFileFlags)
			if err != nil {
				slog.Error("Failed to parse values from files", "error", err)
				return fmt.Errorf("failed to parse values from files: %w", err)
			}
		}

		// Override with CLI values
		for _, value := range valueFlags {
			splitString := strings.SplitN(value, "=", 2)
			switch len(splitString) {
			case 0:
				slog.Error("Empty value flag")
				return fmt.Errorf("empty value flag")
			case 1:
				slog.Error("No value set for value keypair", "value", value)
				return fmt.Errorf("no value set for value keypair: %s", value)
			case 2:
				values[splitString[0]] = splitString[1]
			default:
				slog.Error("Invalid value flag", "value", value)
				return fmt.Errorf("invalid value flag: %s", value)
			}
		}

		// Create logger based on verbose flag
		var loggerLevel slog.Level
		if verboseFlag {
			loggerLevel = slog.LevelDebug
		} else {
			loggerLevel = slog.LevelInfo
		}
		loggerOpts := &slog.HandlerOptions{
			Level: loggerLevel,
		}
		temingoLogger := slog.New(slog.NewTextHandler(os.Stdout, loggerOpts))

		engine := temingo.Engine{
			InputDir:                inputDirFlag,
			OutputDir:               outputDirFlag,
			TemingoignorePath:       temingoignoreFlag,
			TemplateExtension:       templateExtensionFlag,
			MetaTemplateExtension:   metaTemplateExtensionFlag,
			PartialExtension:        partialExtensionFlag,
			MetaFilename:            metaFilenameFlag,
			MarkdownContentFilename: markdownFilenameFlag,
			Values:                  values,
			ValuesFilePaths:         valuesFileFlags,
			NoDeleteOutputDir:       noDeleteOutputDirFlag,
			Verbose:                 verboseFlag,
			DryRun:                  dryRunFlag,
			Logger:                  temingoLogger,
		}

		// Get current directory to pass as targetDir
		cwd, err := os.Getwd()
		if err != nil {
			slog.Error("Failed to get current directory", "error", err)
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		err = engine.InitProject(projectType, cwd)
		if err != nil {
			slog.Error("Failed to initialize project", "error", err)
			return fmt.Errorf("failed to initialize project: %w", err)
		}

		return nil
	},
}
