package refcheck

import (
	"path"
	"strings"
)

// AllowEntry accepts findings for URLs matching a pattern. Checks names the
// categories accepted; empty accepts all of them.
type AllowEntry struct {
	URL    string     `yaml:"url"`
	Checks []Category `yaml:"checks"`
}

// Allowlist is the configured set of accepted findings.
type Allowlist []AllowEntry

// Allows reports whether a finding of category c against rawURL is accepted.
func (a Allowlist) Allows(rawURL string, c Category) bool {
	for _, e := range a {
		if !matchURL(e.URL, rawURL) {
			continue
		}
		if len(e.Checks) == 0 {
			return true
		}
		for _, allowed := range e.Checks {
			if allowed == c {
				return true
			}
		}
	}
	return false
}

// AllowsEverything reports whether every category is accepted for rawURL. When
// it is, the URL need never be requested: the allowlist elides the request
// rather than filtering its result.
func (a Allowlist) AllowsEverything(rawURL string) bool {
	for _, e := range a {
		if matchURL(e.URL, rawURL) && len(e.Checks) == 0 {
			return true
		}
	}
	return false
}

// Filter drops findings the allowlist accepts.
func (a Allowlist) Filter(fs []Finding) []Finding {
	kept := make([]Finding, 0, len(fs))
	for _, f := range fs {
		if !a.Allows(f.Ref.URL, f.Category) {
			kept = append(kept, f)
		}
	}
	return kept
}

// matchURL matches by glob so a versioned URL does not need a fresh entry each
// time its version changes.
//
// A trailing * matches any suffix, including further path segments, because an
// entry written to cover a host is meant to cover everything under it. An
// interior * matches within one path segment, which is what makes a pattern
// like https://cdn.example/lib/*/x.js pin the filename while accepting any
// version.
//
// The trailing form only applies where the prefix ends at a path boundary.
// Without that restriction "https://cdn.example.com*" would also accept
// "https://cdn.example.com.attacker.test/x", silently suppressing findings for -
// and eliding requests to - an entirely different host.
func matchURL(pattern, rawURL string) bool {
	if pattern == "" {
		return false
	}
	if pattern == rawURL {
		return true
	}

	if prefix, ok := strings.CutSuffix(pattern, "*"); ok {
		if strings.HasSuffix(prefix, "/") && strings.HasPrefix(rawURL, prefix) {
			return true
		}
	}

	// Only consult the glob matcher when the pattern actually carries a wildcard.
	// A URL pasted verbatim commonly contains ? and [, which path.Match would
	// otherwise read as metacharacters - ? as any-single-character, and [ as an
	// unterminated class that makes every comparison fail.
	if !strings.ContainsAny(pattern, "*") {
		return false
	}
	matched, err := path.Match(pattern, rawURL)
	return err == nil && matched
}
