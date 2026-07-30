package refcheck

import "testing"

func TestFindingString(t *testing.T) {
	tests := []struct {
		name     string
		finding  Finding
		expected string
	}{
		{
			name: "includes file, role, url, category and reason",
			finding: Finding{
				Ref:      Reference{File: "index.html", URL: "https://x.dev/a", Role: "a href"},
				Category: CategoryStatus,
				Reason:   "responded 404",
			},
			expected: "index.html: a href https://x.dev/a: status: responded 404",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.finding.String(); got != test.expected {
				t.Errorf("String() = %q, want %q", got, test.expected)
			}
		})
	}
}

func TestSortFindings(t *testing.T) {
	findings := []Finding{
		{Ref: Reference{File: "c.html", URL: "https://x.dev/b"}},
		{Ref: Reference{File: "a.html", URL: "https://x.dev/z"}},
		{Ref: Reference{File: "a.html", URL: "https://x.dev/a"}},
	}

	SortFindings(findings)

	// Sorted by file then URL, so output is diffable across builds despite the
	// random map iteration order upstream.
	want := []string{"a.html|https://x.dev/a", "a.html|https://x.dev/z", "c.html|https://x.dev/b"}
	for i, w := range want {
		got := findings[i].Ref.File + "|" + findings[i].Ref.URL
		if got != w {
			t.Errorf("findings[%d] = %q, want %q", i, got, w)
		}
	}
}
