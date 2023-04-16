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
		Verbose:                 false,
		DryRun:                  false,
		Beautify:                false,
		Minify:                  false,
	}
}
