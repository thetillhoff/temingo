package temingo

var (
	// Setting defaults
	verbose bool = false

	inputDir              string
	outputDir             string
	temingoignorePath     string
	templateExtension     string
	metaTemplateExtension string
	componentExtension    string

	defaultTemplateExtension     string = ".template"
	defaultMetaTemplateExtension string = ".metatemplate"
	defaultComponentExtension    string = ".component"
	projectTypes                        = []string{"test", "example", "website"}
)
