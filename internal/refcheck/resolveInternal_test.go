package refcheck

import "testing"

func TestResolveInternal(t *testing.T) {
	outputs := map[string]bool{
		"index.html":             true,
		"style.css":              true,
		"slides/slides.css":      true,
		"slides/index.html":      true,
		"blog/a-post/index.html": true,
		"about.html":             true,
		"images/a.jpg":           true,
	}

	tests := []struct {
		name     string
		refs     []Reference
		wantURLs []string // URLs expected to produce a missing-target finding
	}{
		{
			name:     "root-relative file that exists",
			refs:     []Reference{{File: "index.html", URL: "/slides/slides.css", Origin: OriginInternal}},
			wantURLs: nil,
		},
		{
			name:     "root-relative file that does not exist",
			refs:     []Reference{{File: "index.html", URL: "/slides/slide.css", Origin: OriginInternal}},
			wantURLs: []string{"/slides/slide.css"},
		},
		{
			name:     "directory served by an index document",
			refs:     []Reference{{File: "index.html", URL: "/slides/", Origin: OriginInternal}},
			wantURLs: nil,
		},
		{
			name:     "extensionless path served by a document",
			refs:     []Reference{{File: "index.html", URL: "/about", Origin: OriginInternal}},
			wantURLs: nil,
		},
		{
			name:     "document-relative path",
			refs:     []Reference{{File: "slides/index.html", URL: "slides.css", Origin: OriginInternal}},
			wantURLs: nil,
		},
		{
			name:     "parent-relative path",
			refs:     []Reference{{File: "blog/a-post/index.html", URL: "../../style.css", Origin: OriginInternal}},
			wantURLs: nil,
		},
		{
			name:     "query and fragment are stripped before resolving",
			refs:     []Reference{{File: "index.html", URL: "/style.css?v=1#x", Origin: OriginInternal}},
			wantURLs: nil,
		},
		{
			name:     "unknown directory with no index is reported",
			refs:     []Reference{{File: "index.html", URL: "/reading-notes/", Origin: OriginInternal}},
			wantURLs: []string{"/reading-notes/"},
		},
		{
			name:     "remote references are not resolved here",
			refs:     []Reference{{File: "index.html", URL: "https://x.dev/a", Origin: OriginRemote}},
			wantURLs: nil,
		},
		{
			name:     "escaping the output root is not a finding",
			refs:     []Reference{{File: "index.html", URL: "../../../etc/passwd", Origin: OriginInternal}},
			wantURLs: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ResolveInternal(test.refs, outputs)
			if len(got) != len(test.wantURLs) {
				t.Fatalf("ResolveInternal() = %+v, want %d findings", got, len(test.wantURLs))
			}
			for i, u := range test.wantURLs {
				if got[i].Ref.URL != u {
					t.Errorf("findings[%d].Ref.URL = %q, want %q", i, got[i].Ref.URL, u)
				}
				if got[i].Category != CategoryMissingTarget {
					t.Errorf("findings[%d].Category = %q, want %q", i, got[i].Category, CategoryMissingTarget)
				}
			}
		})
	}
}
