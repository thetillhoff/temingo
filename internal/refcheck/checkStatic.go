package refcheck

// CheckStatic produces every finding derivable from references alone, with no
// network access. A build with no reachability still gets the complete static
// finding set.
func CheckStatic(refs []Reference) []Finding {
	var findings []Finding

	for _, r := range refs {
		// A cross-origin @import cannot be integrity-protected at all: CSS has
		// no syntax for a hash, and an integrity hash on the importing
		// stylesheet does not extend to what it imports. Constraining it needs
		// a CSP style-src rule, or self-hosting the sheet.
		if r.Role == "css @import" && r.Origin == OriginRemote {
			findings = append(findings, Finding{
				Ref: r, Category: CategoryUnverifiedImport,
				Reason: "cross-origin @import cannot carry an integrity hash, and the importing stylesheet's hash does not cover it",
			})
		}

		// Integrity and CORS findings apply only where a browser would verify a
		// hash. Raising them elsewhere is advice the author cannot act on.
		if !r.CanCarryIntegrity || r.Origin != OriginRemote {
			continue
		}

		switch {
		case r.HasIntegrity && !r.HasCrossOrigin:
			findings = append(findings, Finding{
				Ref: r, Category: CategoryMissingCrossOrigin,
				Reason: "integrity hash on a cross-origin subresource with no CORS opt-in; the browser blocks it outright",
			})
		case !r.HasIntegrity:
			findings = append(findings, Finding{
				Ref: r, Category: CategoryMissingIntegrity,
				Reason: "cross-origin subresource with no integrity hash",
			})
		}
	}

	return findings
}
