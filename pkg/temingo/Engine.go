package temingo

type Engine struct {
	InputDir              string
	OutputDir             string
	TemingoignorePath     string
	TemplateExtension     string
	MetaTemplateExtension string
	ComponentExtension    string
	Verbose               bool
}

func DefaultEngine() Engine {
	return Engine{
		InputDir:              "src/",
		OutputDir:             "output/",
		TemingoignorePath:     ".temingoignore",
		TemplateExtension:     ".template",
		MetaTemplateExtension: ".metatemplate",
		ComponentExtension:    ".component",
		Verbose:               false,
	}
}
