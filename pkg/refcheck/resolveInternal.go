package refcheck

import (
	"net/url"
	"path"
	"strings"
)

// ResolveInternal reports references addressing the build's own output that
// resolve to nothing the build produced.
//
// outputPaths is the set of slash-separated, output-root-relative paths the
// build will write. Resolution never touches the filesystem, so it holds under
// a dry run, and it reports only proven absence: anything a server rewrite
// could plausibly satisfy produces no finding.
func ResolveInternal(refs []Reference, outputPaths map[string]bool) []Finding {
	var findings []Finding

	for _, r := range refs {
		if r.Origin != OriginInternal {
			continue
		}

		target, ok := resolveTarget(r)
		if !ok {
			continue
		}

		// A static tree can serve a path as a file, as a directory holding an
		// index document, or as a document whose extension was omitted. A
		// generator cannot know which form a given server prefers, so any of
		// them resolving is enough.
		candidates := []string{
			target,
			path.Join(target, "index.html"),
			target + ".html",
		}
		found := false
		for _, c := range candidates {
			if outputPaths[c] {
				found = true
				break
			}
		}
		if found {
			continue
		}

		findings = append(findings, Finding{
			Ref: r, Category: CategoryMissingTarget,
			Reason: "no file in the build output answers this path",
		})
	}

	return findings
}

// resolveTarget turns a reference into an output-root-relative path. It reports
// false when the reference cannot be resolved to one, in which case no finding
// may be raised.
func resolveTarget(r Reference) (string, bool) {
	u, err := url.Parse(r.URL)
	if err != nil {
		return "", false
	}
	p, err := url.PathUnescape(u.Path)
	if err != nil || p == "" {
		return "", false
	}

	var target string
	if strings.HasPrefix(p, "/") {
		target = strings.TrimPrefix(p, "/")
	} else {
		target = path.Join(path.Dir(r.File), p)
	}
	target = path.Clean(target)

	// A path that climbs out of the output root is not something the build can
	// speak to.
	if target == ".." || strings.HasPrefix(target, "../") {
		return "", false
	}
	if target == "." || target == "" {
		return "", false
	}

	return target, true
}
