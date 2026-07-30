package refcheck

import "testing"

func TestAllowlist(t *testing.T) {
	list := Allowlist{
		{URL: "https://paywalled.example/*"},
		{URL: "https://redirecting.example/*", Checks: []Category{CategoryRedirect}},
		{URL: "https://cdn.example/lib/*/x.js", Checks: []Category{CategoryMissingIntegrity}},
	}

	tests := []struct {
		name           string
		url            string
		category       Category
		wantAllowed    bool
		wantEverything bool
	}{
		{name: "no entry matches", url: "https://other.example/a", category: CategoryStatus},
		{
			name: "entry without checks allows any category", url: "https://paywalled.example/a",
			category: CategoryStatus, wantAllowed: true, wantEverything: true,
		},
		{
			name: "entry without checks allows a different category too", url: "https://paywalled.example/a",
			category: CategoryGated, wantAllowed: true, wantEverything: true,
		},
		{
			name: "trailing star covers nested paths", url: "https://paywalled.example/a/b/c",
			category: CategoryStatus, wantAllowed: true, wantEverything: true,
		},
		{
			name: "narrowed entry allows the named category", url: "https://redirecting.example/docs",
			category: CategoryRedirect, wantAllowed: true,
		},
		{
			name: "narrowed entry does not allow other categories", url: "https://redirecting.example/docs",
			category: CategoryStatus,
		},
		{
			name: "interior star matches one path segment", url: "https://cdn.example/lib/5.2.1/x.js",
			category: CategoryMissingIntegrity, wantAllowed: true,
		},
		{
			name: "interior star does not match a different file", url: "https://cdn.example/lib/5.2.1/y.js",
			category: CategoryMissingIntegrity,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := list.Allows(test.url, test.category); got != test.wantAllowed {
				t.Errorf("Allows(%q, %q) = %v, want %v", test.url, test.category, got, test.wantAllowed)
			}
			if got := list.AllowsEverything(test.url); got != test.wantEverything {
				t.Errorf("AllowsEverything(%q) = %v, want %v", test.url, got, test.wantEverything)
			}
		})
	}
}

func TestAllowlistFilter(t *testing.T) {
	list := Allowlist{{URL: "https://paywalled.example/*"}}
	findings := []Finding{
		{Ref: Reference{URL: "https://paywalled.example/a"}, Category: CategoryGated},
		{Ref: Reference{URL: "https://other.example/b"}, Category: CategoryStatus},
	}

	got := list.Filter(findings)

	if len(got) != 1 {
		t.Fatalf("Filter() returned %d findings, want 1", len(got))
	}
	if got[0].Ref.URL != "https://other.example/b" {
		t.Errorf("Filter() kept %q, want the unallowed one", got[0].Ref.URL)
	}
}
