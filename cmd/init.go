package cmd

import (
	"log"
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

		engine := temingo.Engine{
			InputDir:              inputDirFlag,
			OutputDir:             outputDirFlag,
			TemingoignorePath:     temingoignoreFlag,
			TemplateExtension:     templateExtensionFlag,
			MetaTemplateExtension: metaTemplateExtensionFlag,
			ComponentExtension:    componentExtensionFlag,
			Verbose:               verboseFlag,
			DryRun:                dryRunFlag,
		}

		err := engine.InitProject(args[0]) // There can only be one argument, as specified by `cobra.ExactArgs(1)`
		if err != nil {
			log.Fatalln(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
