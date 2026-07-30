package refcheck

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCacheFetch(t *testing.T) {
	var hits int64

	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write([]byte("hello"))
	})
	mux.HandleFunc("/nocors", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/moved", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusMovedPermanently)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewCache(nil)

	t.Run("200 with CORS", func(t *testing.T) {
		got := c.Fetch(srv.URL+"/ok", "")
		if got.Err != nil {
			t.Fatalf("Err = %v, want nil", got.Err)
		}
		if got.Status != 200 || !got.AllowsCORS {
			t.Errorf("Status=%d AllowsCORS=%v, want 200 true", got.Status, got.AllowsCORS)
		}
	})

	t.Run("200 without CORS", func(t *testing.T) {
		if got := c.Fetch(srv.URL+"/nocors", ""); got.AllowsCORS {
			t.Errorf("AllowsCORS = true, want false")
		}
	})

	t.Run("404", func(t *testing.T) {
		if got := c.Fetch(srv.URL+"/missing", ""); got.Status != 404 {
			t.Errorf("Status = %d, want 404", got.Status)
		}
	})

	t.Run("redirect reports final url", func(t *testing.T) {
		got := c.Fetch(srv.URL+"/moved", "")
		if got.Status != 301 {
			t.Errorf("Status = %d, want 301", got.Status)
		}
		if got.FinalURL != srv.URL+"/ok" {
			t.Errorf("FinalURL = %q, want %q", got.FinalURL, srv.URL+"/ok")
		}
	})

	t.Run("hash of body", func(t *testing.T) {
		got := c.Fetch(srv.URL+"/ok", "sha384")
		// sha384 of "hello", verified with:
		//   printf 'hello' | openssl dgst -sha384 -binary | openssl base64 -A
		want := "sha384-WeF0h3dEjGnea4ANejO7+5/xtGPkQ1TDVTvNucZm+pASWjx5+QOXvfX2oT3oKGhP"
		if got.Hash != want {
			t.Errorf("Hash = %q, want %q", got.Hash, want)
		}
	})

	t.Run("one request per url regardless of call count", func(t *testing.T) {
		before := atomic.LoadInt64(&hits)
		for i := 0; i < 5; i++ {
			c.Fetch(srv.URL+"/ok", "")
		}
		if after := atomic.LoadInt64(&hits); after != before {
			t.Errorf("server saw %d new requests, want 0 - results must be cached", after-before)
		}
	})

	t.Run("a cached hashed result answers an unhashed request", func(t *testing.T) {
		// This is the common case, not an edge one: sri hashes a URL during
		// render, and the same URL is then checked as a reference. It must not
		// be requested twice.
		fresh := NewCache(nil)
		before := atomic.LoadInt64(&hits)
		fresh.Fetch(srv.URL+"/ok", "sha384")
		fresh.Fetch(srv.URL+"/ok", "")
		if after := atomic.LoadInt64(&hits); after-before != 1 {
			t.Errorf("server saw %d requests, want 1", after-before)
		}
	})

	t.Run("unreachable host is indeterminate", func(t *testing.T) {
		got := c.Fetch("https://this-host-does-not-exist.invalid/x", "")
		if got.Err == nil {
			t.Errorf("Err = nil, want an error for an unresolvable host")
		}
	})

	t.Run("allowlisted url is never requested", func(t *testing.T) {
		allowed := NewCache(Allowlist{{URL: srv.URL + "/*"}})
		before := atomic.LoadInt64(&hits)
		got := allowed.Fetch(srv.URL+"/ok", "")
		if after := atomic.LoadInt64(&hits); after != before {
			t.Errorf("server saw a request for an allowlisted url")
		}
		if got.Status != 0 || got.Err != nil {
			t.Errorf("Fetch() = %+v, want a zero result for an elided request", got)
		}
	})
}
