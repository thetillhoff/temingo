package refcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckRemote(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	})
	mux.HandleFunc("/nocors", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/gated", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	mux.HandleFunc("/moved", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusMovedPermanently)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	tests := []struct {
		name     string
		path     string
		ref      Reference
		wantCat  Category
		wantNone bool
	}{
		{name: "200 is clean", path: "/ok", wantNone: true},
		{name: "404 is breakage", path: "/missing", wantCat: CategoryStatus},
		{name: "403 is gated", path: "/gated", wantCat: CategoryGated},
		{name: "301 reports the target", path: "/moved", wantCat: CategoryRedirect},
		{
			name: "200 without CORS breaks an integrity hash", path: "/nocors",
			ref:     Reference{Role: "script src", CanCarryIntegrity: true, HasIntegrity: true, HasCrossOrigin: true},
			wantCat: CategoryNoCORSHeader,
		},
		{
			name: "200 without CORS is fine with no integrity hash", path: "/nocors",
			ref:      Reference{Role: "img src"},
			wantNone: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ref := test.ref
			ref.File = "index.html"
			ref.URL = srv.URL + test.path
			ref.Origin = OriginRemote
			if ref.Role == "" {
				ref.Role = "a href"
			}

			got := CheckRemote([]Reference{ref}, NewCache(nil))

			if test.wantNone {
				if len(got) != 0 {
					t.Fatalf("CheckRemote() = %+v, want no findings", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("CheckRemote() = %+v, want 1 finding", got)
			}
			if got[0].Category != test.wantCat {
				t.Errorf("Category = %q, want %q", got[0].Category, test.wantCat)
			}
		})
	}
}

func TestCheckRemoteUnreachable(t *testing.T) {
	ref := Reference{File: "index.html", URL: "https://does-not-exist.invalid/x", Role: "a href", Origin: OriginRemote}

	got := CheckRemote([]Reference{ref}, NewCache(nil))

	if len(got) != 1 || got[0].Category != CategoryUnreachable {
		t.Fatalf("CheckRemote() = %+v, want one unreachable finding", got)
	}
}
