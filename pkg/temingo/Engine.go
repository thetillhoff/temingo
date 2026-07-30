package temingo

import (
	"log/slog"
	"os"

	"github.com/thetillhoff/temingo/pkg/refcheck"
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

	// Strict makes any reference finding exit non-zero. It draws no distinction
	// between a definite failure and an indeterminate one: a timeout is as fatal
	// as a 404, and the remedy is to run again.
	Strict bool
	// Allow accepts reference findings for matching URLs.
	Allow refcheck.Allowlist

	// linkCache keeps request outcomes for the life of the engine, so watch-mode
	// rebuilds do not re-request unchanged references.
	linkCache *refcheck.Cache
}

// DefaultEngine returns an engine with default values
func DefaultEngine() Engine {
	level := slog.LevelInfo
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
		Strict:                  false,
		Allow:                   nil,
	}
}
