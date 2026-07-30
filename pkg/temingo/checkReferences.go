package temingo

import (
	"fmt"
	"path"

	"github.com/thetillhoff/temingo/pkg/refcheck"
)

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
		switch path.Ext(p) {
		case ".html":
			refs = append(refs, refcheck.CollectHTML(p, content)...)
		case ".css":
			refs = append(refs, refcheck.CollectCSS(p, content)...)
		}
	}

	findings := refcheck.CheckStatic(refs)
	findings = append(findings, refcheck.ResolveInternal(refs, outputPaths)...)

	if engine.linkCache == nil {
		engine.linkCache = refcheck.NewCache(engine.Allow)
	}
	findings = append(findings, refcheck.CheckRemote(refs, engine.linkCache)...)

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
