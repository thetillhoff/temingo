package temingo

import (
	"log/slog"
	"os"
)

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
	ValuesFilePaths         []string
	NoDeleteOutputDir       bool
	Verbose                 bool
	DryRun                  bool
	Beautify                bool
	Minify                  bool
	Logger                  *slog.Logger
}

// DefaultEngine returns an engine with default values
func DefaultEngine() Engine {
	var level slog.Level = slog.LevelInfo
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

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
		ValuesFilePaths:         []string{},
		NoDeleteOutputDir:       false,
		Verbose:                 false,
		DryRun:                  false,
		Beautify:                false,
		Minify:                  false,
		Logger:                  logger,
	}
}
