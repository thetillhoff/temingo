package temingo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTmplSRI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/gone" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	t.Run("default algorithm is sha384", func(t *testing.T) {
		engine := DefaultEngine()
		got, err := engine.tmplSRI(srv.URL + "/x.js")
		if err != nil {
			t.Fatalf("tmplSRI() error = %v", err)
		}
		if !strings.HasPrefix(got, "sha384-") {
			t.Errorf("tmplSRI() = %q, want a sha384- prefix", got)
		}
	})

	t.Run("explicit algorithm is honoured", func(t *testing.T) {
		engine := DefaultEngine()
		got, err := engine.tmplSRI(srv.URL+"/x.js", "sha512")
		if err != nil {
			t.Fatalf("tmplSRI() error = %v", err)
		}
		if !strings.HasPrefix(got, "sha512-") {
			t.Errorf("tmplSRI() = %q, want a sha512- prefix", got)
		}
	})

	t.Run("unknown algorithm is an error", func(t *testing.T) {
		engine := DefaultEngine()
		if _, err := engine.tmplSRI(srv.URL+"/x.js", "md5"); err == nil {
			t.Errorf("tmplSRI() error = nil, want an error for an unsupported algorithm")
		}
	})

	t.Run("local path is an error", func(t *testing.T) {
		engine := DefaultEngine()
		_, err := engine.tmplSRI("/style.css")
		if err == nil {
			t.Fatalf("tmplSRI() error = nil, want an error - the function is remote-only")
		}
		if !strings.Contains(err.Error(), "remote") {
			t.Errorf("error = %q, want it to explain that only remote URLs are supported", err)
		}
	})

	t.Run("unreachable target is a hard error", func(t *testing.T) {
		engine := DefaultEngine()
		if _, err := engine.tmplSRI("https://does-not-exist.invalid/x.js"); err == nil {
			t.Errorf("tmplSRI() error = nil, want an error - there is no correct output without a hash")
		}
	})

	t.Run("error status is a hard error", func(t *testing.T) {
		engine := DefaultEngine()
		if _, err := engine.tmplSRI(srv.URL + "/gone"); err == nil {
			t.Errorf("tmplSRI() error = nil, want an error for a 404")
		}
	})
}
