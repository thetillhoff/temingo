package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thetillhoff/temingo/pkg/temingo"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:       "init {" + strings.Join(temingo.ProjectTypes(), ",") + "}",
	Short:     "Initializes the current directory with an example project. Available types are " + strings.Join(temingo.ProjectTypes(), ", ") + ".",
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: temingo.ProjectTypes(),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err    error
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
				os.Exit(1)
			}
		}

		// Override with CLI values
		for _, value := range valueFlags {
			splitString := strings.SplitN(value, "=", 2)
			switch len(splitString) {
			case 0:
				slog.Error("Empty value flag")
				os.Exit(1)
			case 1:
				slog.Error("No value set for value keypair", "value", value)
				os.Exit(1)
			case 2:
				values[splitString[0]] = splitString[1]
			default:
				slog.Error("Invalid value flag", "value", value)
				os.Exit(1)
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

		err = engine.InitProject(args[0]) // There can only be one argument, as specified by `cobra.ExactArgs(1)`
		if err != nil {
			slog.Error("Failed to initialize project", "error", err)
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
