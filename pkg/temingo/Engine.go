package temingo

type Engine struct {
	InputDir                string
	OutputDir               string
	TemingoignorePath       string
	TemplateExtension       string
	MetaTemplateExtension   string
	PartialExtension        string
	MetaFilename            string
	MarkdownContentFilename string
	Values                  map[string]string
	NoDeleteOutputDir       bool
	Verbose                 bool
	DryRun                  bool
	Beautify                bool
	Minify                  bool
}

// Returns an engine with default values
func DefaultEngine() Engine {
	return Engine{
		InputDir:                "src/",
		OutputDir:               "output/",
		TemingoignorePath:       ".temingoignore",
		TemplateExtension:       ".template",
		MetaTemplateExtension:   ".metatemplate",
		PartialExtension:        ".partial",
		MetaFilename:            "meta.yaml",
		MarkdownContentFilename: "content.md",
		Values:                  map[string]string{},
		NoDeleteOutputDir:       false,
		Verbose:                 false,
		DryRun:                  false,
		Beautify:                false,
		Minify:                  false,
	}
}
