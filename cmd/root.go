package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/thetillhoff/fileIO"
	"github.com/thetillhoff/serve/pkg/serve"
	"github.com/thetillhoff/temingo/pkg/temingo"
	"github.com/urfave/cli/v3"
)

var version = "dev" // This is just the default. The actual value is injected at compiletime

// Execute runs the CLI application
func Execute() {
	// Version flag: only long form (--version) to avoid conflict with -v (verbose)
	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "prints just the version of temingo",
		// No Aliases field = only accepts --version, not -v (which is used for verbose)
	}
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Println(cmd.Root().Version)
	}

	app := &cli.Command{
		Name:    "temingo",
		Usage:   "A template engine for static site generation",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "config file (default is .temingo.yaml in the current directory)",
				Sources: cli.EnvVars("TEMINGO_CONFIG"),
			},
			&cli.StringFlag{
				Name:    "inputDir",
				Aliases: []string{"i"},
				Usage:   "inputDir contains the source files",
				Value:   "src/",
				Sources: cli.EnvVars("TEMINGO_INPUT_DIR"),
			},
			&cli.StringFlag{
				Name:    "outputDir",
				Aliases: []string{"o"},
				Usage:   "outputDir is where temingo builds to",
				Value:   "output/",
				Sources: cli.EnvVars("TEMINGO_OUTPUT_DIR"),
			},
			&cli.StringFlag{
				Name:    "temingoignore",
				Usage:   "path to the temingo ignore file (works like gitignore)",
				Value:   ".temingoignore",
				Sources: cli.EnvVars("TEMINGO_IGNORE"),
			},
			&cli.StringFlag{
				Name:    "templateExtension",
				Aliases: []string{"t"},
				Usage:   "templateExtension marks a file as template that correlates to a rendered file",
				Value:   ".template",
				Sources: cli.EnvVars("TEMINGO_TEMPLATE_EXT"),
			},
			&cli.StringFlag{
				Name:    "metaTemplateExtension",
				Aliases: []string{"m"},
				Usage:   "metaTemplateExtension marks a file as template that correlates to multiple rendered files",
				Value:   ".metatemplate",
				Sources: cli.EnvVars("TEMINGO_META_TEMPLATE_EXT"),
			},
			&cli.StringFlag{
				Name:    "partialExtension",
				Aliases: []string{"c"},
				Usage:   "partialExtension marks a file as partial template without a rendered file",
				Value:   ".partial",
				Sources: cli.EnvVars("TEMINGO_PARTIAL_EXT"),
			},
			&cli.StringFlag{
				Name:    "metaFilename",
				Usage:   "the yaml files for the metadata",
				Value:   "meta.yaml",
				Sources: cli.EnvVars("TEMINGO_META_FILENAME"),
			},
			&cli.StringFlag{
				Name:    "markdownFilename",
				Usage:   "the markdown files for the markdown contents",
				Value:   "content.md",
				Sources: cli.EnvVars("TEMINGO_MARKDOWN_FILENAME"),
			},
			&cli.StringSliceFlag{
				Name:    "value",
				Usage:   "value for the templates (`key=value`), multiple occurrences are possible",
				Sources: cli.EnvVars("TEMINGO_VALUE"),
			},
			&cli.StringSliceFlag{
				Name:    "valuesfile",
				Usage:   "path to a YAML file containing key-value pairs for the templates (can be specified multiple times, files are merged)",
				Sources: cli.EnvVars("TEMINGO_VALUES_FILE"),
			},
			&cli.BoolFlag{
				Name:    "noDeleteOutputDir",
				Usage:   "don't delete the outputDir before building",
				Sources: cli.EnvVars("TEMINGO_NO_DELETE_OUTPUT_DIR"),
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbose increases the level of detail of the logs",
				Sources: cli.EnvVars("TEMINGO_VERBOSE"),
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Usage:   "don't output files",
				Sources: cli.EnvVars("TEMINGO_DRY_RUN"),
			},
			&cli.BoolFlag{
				Name:    "watch",
				Aliases: []string{"w"},
				Usage:   "watch makes temingo continiously watch for filesystem changes",
				Sources: cli.EnvVars("TEMINGO_WATCH"),
			},
			&cli.BoolFlag{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "serve makes temingo serve your outputDir with a small simple webserver",
				Sources: cli.EnvVars("TEMINGO_SERVE"),
			},
		},
		Commands: []*cli.Command{
			initCommand,
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
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
			watchFlag := cmd.Bool("watch")
			serveFlag := cmd.Bool("serve")

			// Load config file if specified
			config, err := loadConfig(cfgFile)
			if err != nil {
				slog.Error("Failed to load config", "error", err)
				return err
			}

			// Apply config values to flags (only if not set via CLI)
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
				return fmt.Errorf("build failed: %w", err)
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
					return fmt.Errorf("failed to start file watcher: %w", err)
				}
			}

			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}
