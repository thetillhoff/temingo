package refcheck

import (
	"regexp"
	"strings"
)

// ponytail: regexes, not a CSS tokenizer. These match every url() and @import
// form seen in practice, but can also match inside a comment or a string
// literal. False positives are silenced by the allowlist. Replace with a real
// tokenizer when one arrives for CSS minification/beautification - the roadmap
// already wants one.
var (
	cssURLRe    = regexp.MustCompile(`url\(\s*(?:"([^"]*)"|'([^']*)'|([^)'"\s]+))`)
	cssImportRe = regexp.MustCompile(`@import\s+(?:url\(\s*)?(?:"([^"]*)"|'([^']*)'|([^)'";\s]+))`)
)

// CollectCSS returns every reference in stylesheet content. No CSS reference can
// ever carry an integrity hash, because CSS has no syntax for one.
func CollectCSS(file string, css []byte) []Reference {
	refs := []Reference{}
	s := string(css)

	// @import first, so an imported URL is reported as an import rather than as
	// the url() it may be wrapped in.
	imports := map[string]bool{}
	for _, m := range cssImportRe.FindAllStringSubmatch(s, -1) {
		u := firstNonEmpty(m[1], m[2], m[3])
		imports[u] = true
		if r, ok := cssRef(file, "css @import", u); ok {
			refs = append(refs, r)
		}
	}

	for _, m := range cssURLRe.FindAllStringSubmatch(s, -1) {
		u := firstNonEmpty(m[1], m[2], m[3])
		if imports[u] {
			continue
		}
		if r, ok := cssRef(file, "css url()", u); ok {
			refs = append(refs, r)
		}
	}

	return refs
}

func cssRef(file, role, rawURL string) (Reference, bool) {
	u := strings.TrimSpace(rawURL)
	origin := Classify(u)
	if origin == OriginIgnored {
		return Reference{}, false
	}
	return Reference{File: file, URL: u, Role: role, Origin: origin}, true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
