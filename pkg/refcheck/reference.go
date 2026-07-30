// Package refcheck collects references from rendered output and reports the
// ones that are broken, unverifiable, or point at nothing the build produced.
package refcheck

import "strings"

// Origin classifies what a reference addresses.
type Origin int

const (
	// OriginInternal addresses the build's own output.
	OriginInternal Origin = iota
	// OriginRemote addresses an http(s) origin.
	OriginRemote
	// OriginIgnored addresses nothing fetchable: a fragment, or a scheme such
	// as mailto or tel.
	OriginIgnored
)

// Reference is one addressable target found in rendered output.
type Reference struct {
	// File is the output-relative path of the file the reference was found in.
	File string
	// URL is the target exactly as written.
	URL string
	// Role names the syntactic position, for reporting back to the author.
	Role string

	Origin Origin

	// CanCarryIntegrity reports whether a browser would verify an integrity
	// hash on this reference. It is a property of the reference, not of its
	// element: a stylesheet reached from inside CSS can never carry one.
	CanCarryIntegrity bool
	HasIntegrity      bool
	HasCrossOrigin    bool
}

// Classify determines what a reference addresses from the URL alone.
func Classify(rawURL string) Origin {
	u := strings.TrimSpace(rawURL)
	if u == "" || strings.HasPrefix(u, "#") {
		return OriginIgnored
	}
	lower := strings.ToLower(u)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return OriginRemote
	}
	// A protocol-relative URL addresses a remote origin.
	if strings.HasPrefix(u, "//") {
		return OriginRemote
	}
	// Any other scheme is not fetchable as a build artifact.
	if i := strings.Index(u, ":"); i > 0 {
		if !strings.ContainsAny(u[:i], "/?#") {
			return OriginIgnored
		}
	}
	return OriginInternal
}
