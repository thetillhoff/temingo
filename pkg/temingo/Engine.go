package temingo

import (
	"log/slog"
	"os"

	"github.com/thetillhoff/temingo/internal/refcheck"
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
	// NoRemoteChecks skips every check that needs a request, leaving the static
	// and internal ones - which need no network - in place. Rendering stays
	// offline-capable unless a template calls sri, which cannot produce a hash
	// without fetching.
	NoRemoteChecks bool
	// AllowInsecureScheme stops reporting references fetched over plain http.
	AllowInsecureScheme bool

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
		NoRemoteChecks:          false,
		AllowInsecureScheme:     false,
	}
}

// The reference checker is internal, but Engine.Allow is an exported field, so
// the three types needed to populate it are aliased here. Without them the field
// would be unsettable from outside this module - an exported field nobody can
// use.

// Allowlist is the configured set of accepted reference findings.
type Allowlist = refcheck.Allowlist

// AllowEntry accepts findings for URLs matching a pattern. An entry naming no
// checks accepts every category for its URL.
type AllowEntry = refcheck.AllowEntry

// Category identifies a kind of reference finding, as named by an AllowEntry.
type Category = refcheck.Category
