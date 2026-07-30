package refcheck

import (
	"strings"
	"testing"
)

func TestCheckInsecureScheme(t *testing.T) {
	tests := []struct {
		name       string
		ref        Reference
		wantFound  bool
		wantReason string // substring, checked when a finding is expected
	}{
		{
			name:      "https link is clean",
			ref:       Reference{URL: "https://example.com/a", Role: "a href", Origin: OriginRemote},
			wantFound: false,
		},
		{
			name:       "http link is reported",
			ref:        Reference{URL: "http://example.com/a", Role: "a href", Origin: OriginRemote},
			wantFound:  true,
			wantReason: "read and altered in transit",
		},
		{
			name: "http subresource is reported as mixed content",
			ref: Reference{URL: "http://cdn.example/x.js", Role: "script src", Origin: OriginRemote,
				CanCarryIntegrity: true},
			wantFound:  true,
			wantReason: "mixed content",
		},
		{
			name:      "localhost is exempt",
			ref:       Reference{URL: "http://localhost:3000/x", Role: "a href", Origin: OriginRemote},
			wantFound: false,
		},
		{
			name:      "loopback ipv4 is exempt",
			ref:       Reference{URL: "http://127.0.0.1:8080/x", Role: "a href", Origin: OriginRemote},
			wantFound: false,
		},
		{
			name:      "loopback ipv6 is exempt",
			ref:       Reference{URL: "http://[::1]:8080/x", Role: "a href", Origin: OriginRemote},
			wantFound: false,
		},
		{
			name:      "uppercase scheme is still http",
			ref:       Reference{URL: "HTTP://example.com/a", Role: "a href", Origin: OriginRemote},
			wantFound: true,
		},
		{
			name:      "internal path is not a scheme problem",
			ref:       Reference{URL: "/style.css", Role: "link stylesheet href", Origin: OriginInternal},
			wantFound: false,
		},
		{
			name: "protocol-relative url inherits the page scheme and is not reported",
			ref:  Reference{URL: "//cdn.example/x.js", Role: "script src", Origin: OriginRemote},
			// On an https page this is https. Reporting it would be wrong.
			wantFound: false,
		},
		{
			name:      "a non-loopback private address is still insecure",
			ref:       Reference{URL: "http://192.168.1.10/x", Role: "a href", Origin: OriginRemote},
			wantFound: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CheckInsecureScheme([]Reference{test.ref})

			if !test.wantFound {
				if len(got) != 0 {
					t.Fatalf("CheckInsecureScheme() = %+v, want no findings", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("CheckInsecureScheme() = %+v, want 1 finding", got)
			}
			if got[0].Category != CategoryInsecureScheme {
				t.Errorf("Category = %q, want %q", got[0].Category, CategoryInsecureScheme)
			}
			if test.wantReason != "" && !strings.Contains(got[0].Reason, test.wantReason) {
				t.Errorf("Reason = %q, want it to contain %q", got[0].Reason, test.wantReason)
			}
		})
	}
}
