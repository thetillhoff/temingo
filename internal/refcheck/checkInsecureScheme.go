package refcheck

import (
	"net"
	"net/url"
	"strings"
)

// CheckInsecureScheme reports references fetched over plain http.
//
// It needs no network access: the scheme is visible in the reference itself.
// Loopback targets are exempt, because a local development server legitimately
// serves plain http and there is no network for anyone to sit on.
func CheckInsecureScheme(refs []Reference) []Finding {
	var findings []Finding

	for _, r := range refs {
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.URL)), "http://") {
			continue
		}
		if isLoopbackURL(r.URL) {
			continue
		}

		reason := "fetched over plain http, so it can be read and altered in transit"
		if r.CanCarryIntegrity {
			// A browser refuses this outright on an https page, so it is breakage
			// rather than advice.
			reason = "subresource fetched over plain http; a browser blocks it as mixed content on an https page"
		}

		findings = append(findings, Finding{
			Ref: r, Category: CategoryInsecureScheme, Reason: reason,
		})
	}

	return findings
}

// isLoopbackURL reports whether a URL addresses this machine.
func isLoopbackURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	host := u.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}

	return false
}
