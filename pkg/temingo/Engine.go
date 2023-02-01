package temingo

type Engine struct {
	InputDir              string
	OutputDir             string
	TemingoignorePath     string
	TemplateExtension     string
	MetaTemplateExtension string
	ComponentExtension    string
	MetaFilename          string
	Verbose               bool
	DryRun                bool
}

// Returns an engine with default values
func DefaultEngine() Engine {
	return Engine{
		InputDir:              "src/",
		OutputDir:             "output/",
		TemingoignorePath:     ".temingoignore",
		TemplateExtension:     ".template",
		MetaTemplateExtension: ".metatemplate",
		ComponentExtension:    ".component",
		MetaFilename:          "meta.yaml",
		Verbose:               false,
		DryRun:                false,
	}
}
