package temingo

import (
	"fmt"
	"os"
	"path"

	"github.com/thetillhoff/temingo/internal/refcheck"
)

// collectFrom returns the references in one output file, dispatching on its
// extension.
func (engine *Engine) collectFrom(outputPath string, content []byte) []refcheck.Reference {
	switch path.Ext(outputPath) {
	case ".html":
		return refcheck.CollectHTML(outputPath, content)
	case ".css":
		return refcheck.CollectCSS(outputPath, content)
	default:
		return nil
	}
}

// checkReferences collects every reference in the rendered output and reports
// the ones that are broken, unverifiable, or address nothing the build
// produced.
//
// It runs on the in-memory rendered content and the set of paths the build will
// write, so it needs no filesystem reads and holds under a dry run. Under
// Strict, any finding is returned as an error so the process exits non-zero.
func (engine *Engine) checkReferences(rendered map[string][]byte, staticPaths []string) error {
	var refs []refcheck.Reference

	outputPaths := make(map[string]bool, len(rendered)+len(staticPaths))
	for p := range rendered {
		outputPaths[path.Clean(p)] = true
	}
	for _, p := range staticPaths {
		outputPaths[path.Clean(p)] = true
	}

	for p, content := range rendered {
		refs = append(refs, engine.collectFrom(p, content)...)
	}

	// Static files are copied verbatim rather than rendered, so their references
	// are only visible by reading the source. Skipping them would leave every
	// hand-written page and stylesheet unchecked - which is most stylesheets.
	for _, p := range staticPaths {
		if ext := path.Ext(p); ext != ".html" && ext != ".css" {
			continue
		}
		content, err := os.ReadFile(path.Join(engine.InputDir, p))
		if err != nil {
			engine.Logger.Debug("Skipping unreadable static file during reference check", "path", p, "error", err)
			continue
		}
		refs = append(refs, engine.collectFrom(p, content)...)
	}

	findings := refcheck.CheckStatic(refs)

	if !engine.AllowInsecureScheme {
		findings = append(findings, refcheck.CheckInsecureScheme(refs)...)
	}

	// With NoDeleteOutputDir the output tree also holds whatever earlier builds
	// or other tooling left there, and none of that is in outputPaths. Absence
	// from the set would prove nothing, so resolution is skipped entirely rather
	// than reporting files that are present on disk.
	if engine.NoDeleteOutputDir {
		engine.Logger.Debug("Skipping internal reference resolution because noDeleteOutputDir is set")
	} else {
		findings = append(findings, refcheck.ResolveInternal(refs, outputPaths)...)
	}

	if engine.NoRemoteChecks {
		engine.Logger.Debug("Skipping remote reference checks because noRemoteChecks is set")
	} else {
		if engine.linkCache == nil {
			engine.linkCache = refcheck.NewCache(engine.Allow)
		}
		findings = append(findings, refcheck.CheckRemote(refs, engine.linkCache)...)
	}

	findings = engine.Allow.Filter(findings)
	refcheck.SortFindings(findings)

	for _, f := range findings {
		engine.Logger.Warn("Reference finding",
			"file", f.Ref.File,
			"url", f.Ref.URL,
			"role", f.Ref.Role,
			"category", string(f.Category),
			"reason", f.Reason,
		)
	}

	if engine.Strict && len(findings) > 0 {
		return fmt.Errorf("%d reference findings, and strict mode is enabled", len(findings))
	}

	return nil
}
