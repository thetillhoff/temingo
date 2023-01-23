package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
	"github.com/thetillhoff/temingo/pkg/fileIO"
	"github.com/thetillhoff/temingo/pkg/temingo"

	"github.com/spf13/viper"
)

var (
	cfgFile                   string
	inputDirFlag              string
	outputDirFlag             string
	temingoignoreFlag         string
	templateExtensionFlag     string
	metaTemplateExtensionFlag string
	componentExtensionFlag    string

	verboseFlag bool
	watchFlag   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "temingo",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.NoArgs,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err error
		)

		engine := temingo.Engine{
			InputDir:              inputDirFlag,
			OutputDir:             outputDirFlag,
			TemingoignorePath:     temingoignoreFlag,
			TemplateExtension:     templateExtensionFlag,
			MetaTemplateExtension: metaTemplateExtensionFlag,
			ComponentExtension:    componentExtensionFlag,
			Verbose:               verboseFlag,
		}

		if watchFlag {
			log.Println("*** Started to watch for file changes ***")

			err = fileIO.Watch(
				[]string{
					engine.InputDir,
					engine.TemingoignorePath,
				},
				[]string{
					engine.OutputDir,
					".git",
				},
				engine.Verbose,
				100*time.Millisecond,
				func(event watcher.Event) error {
					log.Println("*** Rebuild triggered by a change detected in", event.Path, "***")
					// TODO inform frontend via websocket connection
					return engine.Render()
				})
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			err = engine.Render()
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("Build complete.")
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
	rootCmd.PersistentFlags().StringVarP(&componentExtensionFlag, "componentExtension", "c", ".component", "componentExtension marks a file as partial template without a rendered file")

	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose increases the level of detail of the logs")
	rootCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "watch makes temingo continiously watch for filesystem changes")
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
