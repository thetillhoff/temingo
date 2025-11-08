package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
	"github.com/thetillhoff/fileIO"
	"github.com/thetillhoff/serve/pkg/serve"
	"github.com/thetillhoff/temingo/pkg/temingo"

	"github.com/spf13/viper"
)

var (
	cfgFile string

	// inputDir must end with a slash
	inputDirFlag string

	// outputDir must end with a slash
	outputDirFlag string

	temingoignoreFlag         string
	templateExtensionFlag     string
	metaTemplateExtensionFlag string
	partialExtensionFlag      string
	metaFilenameFlag          string
	markdownFilenameFlag      string
	valueFlags                []string
	valuesFileFlags           []string

	verboseFlag           bool
	dryRunFlag            bool
	noDeleteOutputDirFlag bool
	watchFlag             bool
	serveFlag             bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:  "temingo",
	Args: cobra.NoArgs,
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

		temingoEngine := temingo.Engine{
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
			Beautify:                true,
			Minify:                  false,
			Logger:                  temingoLogger,
		}

		// Build once
		err = temingoEngine.Render()
		if err != nil {
			slog.Error("Build failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Build complete")

		if serveFlag { // Start webserver if desired
			serveEngine := serve.DefaultEngine()
			serveEngine.Ipaddress = "127.0.0.1" // Only listen to local connections
			serveEngine.Directory = outputDirFlag
			serveEngine.Verbose = verboseFlag
			go func() { // Start the webserver in the background
				err = serveEngine.Serve()
				if err != nil {
					slog.Error("Failed to start webserver", "error", err)
					os.Exit(1)
				}
			}()
		}

		if watchFlag { // Start watching if desired
			slog.Info("Started to watch for file changes")

			err = fileIO.Watch(
				[]string{
					temingoEngine.InputDir,
					temingoEngine.TemingoignorePath,
				},
				[]string{
					temingoEngine.OutputDir,
					".git",
				},
				temingoEngine.Verbose,
				100*time.Millisecond,
				func(event watcher.Event) error {
					slog.Info("Rebuild triggered by file change", "path", event.Path)
					// TODO inform frontend via websocket connection
					err = temingoEngine.Render()
					if err != nil {
						slog.Error("Rebuild failed", "error", err) // Print errors when in watch mode
					}
					return nil // Ignore errors on Rendering when in watch mode (apart from printing them)
				})
			if err != nil {
				slog.Error("Failed to start file watcher", "error", err)
				os.Exit(1)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.temingo.yaml)")

	rootCmd.PersistentFlags().StringVarP(&inputDirFlag, "inputDir", "i", "src/", "inputDir contains the source files")
	rootCmd.PersistentFlags().StringVarP(&outputDirFlag, "outputDir", "o", "output/", "outputDir is where temingo builds to")
	rootCmd.PersistentFlags().StringVar(&temingoignoreFlag, "temingoignore", ".temingoignore", "path to the temingo ignore file (works like gitignore`)")
	rootCmd.PersistentFlags().StringVarP(&templateExtensionFlag, "templateExtension", "t", ".template", "templateExtension marks a file as template that correlates to a rendered file")
	rootCmd.PersistentFlags().StringVarP(&metaTemplateExtensionFlag, "metaTemplateExtension", "m", ".metatemplate", "metaTemplateExtension marks a file as template that correlates to multiple rendered files")
	rootCmd.PersistentFlags().StringVarP(&partialExtensionFlag, "partialExtension", "c", ".partial", "partialExtension marks a file as partial template without a rendered file")
	rootCmd.PersistentFlags().StringVar(&metaFilenameFlag, "metaFilename", "meta.yaml", "the yaml files for the metadata")
	rootCmd.PersistentFlags().StringVar(&markdownFilenameFlag, "markdownFilename", "content.md", "the markdown files for the markdown contents")
	rootCmd.PersistentFlags().StringSliceVar(&valueFlags, "value", []string{}, "value for the templates (`key=value`), multiple occurrences are possible")
	rootCmd.PersistentFlags().StringSliceVar(&valuesFileFlags, "valuesfile", []string{}, "path to a YAML file containing key-value pairs for the templates (can be specified multiple times, files are merged)")

	rootCmd.PersistentFlags().BoolVar(&noDeleteOutputDirFlag, "noDeleteOutputDir", false, "don't delete the outputDir before building")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose increases the level of detail of the logs")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "don't output files")
	rootCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "watch makes temingo continiously watch for filesystem changes")
	rootCmd.Flags().BoolVarP(&serveFlag, "serve", "s", false, "serve makes temingo serve your outputDir with a small simple webserver")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".temingo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".temingo")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
