package refcheck

import (
	"fmt"
	"net/http"
)

// CheckRemote requests every remote reference and reports what it finds. Each
// distinct URL is requested at most once, however many references share it.
//
// ponytail: sequential. The cache lives for the process, so only the first
// build in a watch session pays. Fan out here if that becomes too slow.
func CheckRemote(refs []Reference, c *Cache) []Finding {
	var findings []Finding

	for _, r := range refs {
		if r.Origin != OriginRemote {
			continue
		}

		result := c.Fetch(r.URL, "")

		switch {
		case result.Err != nil:
			findings = append(findings, Finding{
				Ref: r, Category: CategoryUnreachable,
				Reason: fmt.Sprintf("could not be determined: %v", result.Err),
			})
			continue
		case result.Status == 0:
			// The request was elided by the allowlist.
			continue
		case result.Status == http.StatusUnauthorized || result.Status == http.StatusForbidden:
			findings = append(findings, Finding{
				Ref: r, Category: CategoryGated,
				Reason: fmt.Sprintf("responded %d; expected for paywalled or login-gated targets", result.Status),
			})
		case result.Status >= 400:
			findings = append(findings, Finding{
				Ref: r, Category: CategoryStatus,
				Reason: fmt.Sprintf("responded %d", result.Status),
			})
		case result.Status >= 300:
			findings = append(findings, Finding{
				Ref: r, Category: CategoryRedirect,
				Reason: fmt.Sprintf("responded %d, redirecting to %s; replace the reference with that target", result.Status, result.FinalURL),
			})
		}

		// An integrity hash the browser cannot verify for lack of readable
		// bytes is breakage, not advice.
		if r.CanCarryIntegrity && r.HasIntegrity && result.Status < 300 && !result.AllowsCORS {
			findings = append(findings, Finding{
				Ref: r, Category: CategoryNoCORSHeader,
				Reason: "carries an integrity hash but the target sends no Access-Control-Allow-Origin, so the browser cannot verify it and blocks the subresource",
			})
		}
	}

	return findings
}
