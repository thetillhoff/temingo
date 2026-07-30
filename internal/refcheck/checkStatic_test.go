package refcheck

import "testing"

func TestCheckStatic(t *testing.T) {
	tests := []struct {
		name     string
		refs     []Reference
		wantCats []Category
	}{
		{
			name: "cross-origin script without integrity",
			refs: []Reference{{
				File: "i.html", URL: "https://cdn.example/x.js", Role: "script src",
				Origin: OriginRemote, CanCarryIntegrity: true,
			}},
			wantCats: []Category{CategoryMissingIntegrity},
		},
		{
			name: "integrity without crossorigin is breakage",
			refs: []Reference{{
				File: "i.html", URL: "https://cdn.example/x.js", Role: "script src",
				Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true,
			}},
			wantCats: []Category{CategoryMissingCrossOrigin},
		},
		{
			name: "integrity with crossorigin is clean",
			refs: []Reference{{
				File: "i.html", URL: "https://cdn.example/x.js", Role: "script src",
				Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true,
			}},
			wantCats: nil,
		},
		{
			name: "same-origin script needs no integrity",
			refs: []Reference{{
				File: "i.html", URL: "/local.js", Role: "script src",
				Origin: OriginInternal, CanCarryIntegrity: true,
			}},
			wantCats: nil,
		},
		{
			name: "image is never asked for integrity",
			refs: []Reference{{
				File: "i.html", URL: "https://cdn.example/a.jpg", Role: "img src",
				Origin: OriginRemote,
			}},
			wantCats: nil,
		},
		{
			name: "css reference is never asked for integrity",
			refs: []Reference{{
				File: "s.css", URL: "https://cdn.example/f.woff2", Role: "css url()",
				Origin: OriginRemote,
			}},
			wantCats: nil,
		},
		{
			name: "cross-origin import is inherently unprotectable",
			refs: []Reference{{
				File: "a.css", URL: "https://other.example/b.css", Role: "css @import",
				Origin: OriginRemote,
			}},
			wantCats: []Category{CategoryUnverifiedImport},
		},
		{
			name: "same-origin import is fine",
			refs: []Reference{{
				File: "a.css", URL: "b.css", Role: "css @import",
				Origin: OriginInternal,
			}},
			wantCats: nil,
		},
		{
			name: "a verified cross-origin stylesheet is not itself an import finding",
			refs: []Reference{{
				File: "i.html", URL: "https://cdn.example/a.css", Role: "link stylesheet href",
				Origin: OriginRemote, CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true,
			}},
			wantCats: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CheckStatic(test.refs)
			if len(got) != len(test.wantCats) {
				t.Fatalf("CheckStatic() returned %d findings (%+v), want %d", len(got), got, len(test.wantCats))
			}
			for i, c := range test.wantCats {
				if got[i].Category != c {
					t.Errorf("findings[%d].Category = %q, want %q", i, got[i].Category, c)
				}
			}
		})
	}
}
