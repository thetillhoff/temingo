package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thetillhoff/temingo/pkg/temingo"
)

var (
	projectTypeFlag string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		engine := temingo.Engine{
			InputDir:              inputDirFlag,
			OutputDir:             outputDirFlag,
			TemingoignorePath:     temingoignoreFlag,
			TemplateExtension:     templateExtensionFlag,
			MetaTemplateExtension: metaTemplateExtensionFlag,
			ComponentExtension:    componentExtensionFlag,
			Verbose:               verboseFlag,
		}

		err := engine.InitProject(projectTypeFlag)
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
	initCmd.Flags().StringVar(&projectTypeFlag, "type", "example", "The type of project for which initial files should be generated (options: "+strings.Join(temingo.ProjectTypes, ", ")+")")
}
