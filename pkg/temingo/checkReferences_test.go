package temingo

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/thetillhoff/temingo/pkg/refcheck"
)

func TestCheckReferencesNoRemoteChecks(t *testing.T) {
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	rendered := map[string][]byte{
		// One remote reference that would 404, and one internal one that is broken.
		"index.html": []byte(`<a href="` + srv.URL + `/gone">x</a><img src="/nope.jpg">`),
	}

	t.Run("remote checks run by default", func(t *testing.T) {
		var buf bytes.Buffer
		engine := DefaultEngine()
		engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))

		if err := engine.checkReferences(rendered, nil); err != nil {
			t.Fatalf("checkReferences() = %v", err)
		}
		if out := buf.String(); !strings.Contains(out, "status") {
			t.Errorf("expected a status finding, got:\n%s", out)
		}
		if atomic.LoadInt64(&hits) == 0 {
			t.Errorf("no request was made")
		}
	})

	t.Run("noRemoteChecks skips requests but keeps offline checks", func(t *testing.T) {
		var buf bytes.Buffer
		engine := DefaultEngine()
		engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))
		engine.NoRemoteChecks = true

		before := atomic.LoadInt64(&hits)
		if err := engine.checkReferences(rendered, nil); err != nil {
			t.Fatalf("checkReferences() = %v", err)
		}
		if after := atomic.LoadInt64(&hits); after != before {
			t.Errorf("server saw %d new requests, want 0", after-before)
		}

		out := buf.String()
		if strings.Contains(out, "category=status") {
			t.Errorf("remote finding reported while remote checks are off:\n%s", out)
		}
		// The internal check needs no network and must still run.
		if !strings.Contains(out, "missing-target") {
			t.Errorf("expected the internal finding to survive, got:\n%s", out)
		}
	})
}

func TestCheckReferencesInsecureScheme(t *testing.T) {
	rendered := map[string][]byte{
		"index.html": []byte(`<a href="http://example.com/a">x</a>`),
	}

	t.Run("reported by default", func(t *testing.T) {
		var buf bytes.Buffer
		engine := DefaultEngine()
		engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))
		engine.NoRemoteChecks = true // isolate: no network needed for this check

		if err := engine.checkReferences(rendered, nil); err != nil {
			t.Fatalf("checkReferences() = %v", err)
		}
		if out := buf.String(); !strings.Contains(out, "insecure-scheme") {
			t.Errorf("expected an insecure-scheme finding, got:\n%s", out)
		}
	})

	t.Run("suppressed by allowInsecureScheme", func(t *testing.T) {
		var buf bytes.Buffer
		engine := DefaultEngine()
		engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))
		engine.NoRemoteChecks = true
		engine.AllowInsecureScheme = true

		if err := engine.checkReferences(rendered, nil); err != nil {
			t.Fatalf("checkReferences() = %v", err)
		}
		if out := buf.String(); strings.Contains(out, "insecure-scheme") {
			t.Errorf("finding reported while allowInsecureScheme is set:\n%s", out)
		}
	})
}

// TestCheckReferencesAppliesAllowlist pins the Engine.Allow wiring, which had no
// end-to-end coverage - only the allowlist type itself was tested.
func TestCheckReferencesAppliesAllowlist(t *testing.T) {
	rendered := map[string][]byte{
		"index.html": []byte(`<img src="/nope.jpg">`),
	}

	var buf bytes.Buffer
	engine := DefaultEngine()
	engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))
	engine.NoRemoteChecks = true

	if err := engine.checkReferences(rendered, nil); err != nil {
		t.Fatalf("checkReferences() = %v", err)
	}
	if !strings.Contains(buf.String(), "missing-target") {
		t.Fatalf("expected a finding without an allowlist, got:\n%s", buf.String())
	}

	buf.Reset()
	engine.Allow = refcheck.Allowlist{{URL: "/nope.jpg"}}

	if err := engine.checkReferences(rendered, nil); err != nil {
		t.Fatalf("checkReferences() = %v", err)
	}
	if strings.Contains(buf.String(), "missing-target") {
		t.Errorf("allowlisted finding was still reported:\n%s", buf.String())
	}
}

func TestCheckReferences(t *testing.T) {
	tests := []struct {
		name        string
		strict      bool
		rendered    map[string][]byte
		staticPaths []string
		wantLogged  []string
		wantAbsent  []string
		wantErr     bool
	}{
		{
			name: "resolvable internal reference is silent",
			rendered: map[string][]byte{
				"index.html": []byte(`<link rel="stylesheet" href="/style.css">`),
			},
			staticPaths: []string{"style.css"},
			wantAbsent:  []string{"missing-target"},
		},
		{
			name: "mistyped internal reference is reported",
			rendered: map[string][]byte{
				"index.html": []byte(`<link rel="stylesheet" href="/styl.css">`),
			},
			staticPaths: []string{"style.css"},
			wantLogged:  []string{"missing-target", "/styl.css"},
		},
		{
			name: "directory reference resolves through an index document",
			rendered: map[string][]byte{
				"index.html":        []byte(`<a href="/slides/">Slides</a>`),
				"slides/index.html": []byte(`<p>decks</p>`),
			},
			wantAbsent: []string{"missing-target"},
		},
		{
			name: "commented-out reference is not reported",
			rendered: map[string][]byte{
				"index.html": []byte(`<nav><!-- <a href="/reading-notes/">Notes</a> --></nav>`),
			},
			wantAbsent: []string{"missing-target", "reading-notes"},
		},
		{
			name: "url in a code element is not reported",
			rendered: map[string][]byte{
				"index.html": []byte(`<code>&lt;a href="/nope/"&gt;x&lt;/a&gt;</code>`),
			},
			wantAbsent: []string{"missing-target"},
		},
		{
			name:   "strict fails the build on a finding",
			strict: true,
			rendered: map[string][]byte{
				"index.html": []byte(`<link rel="stylesheet" href="/styl.css">`),
			},
			staticPaths: []string{"style.css"},
			wantLogged:  []string{"missing-target"},
			wantErr:     true,
		},
		{
			name: "non-strict never fails the build",
			rendered: map[string][]byte{
				"index.html": []byte(`<link rel="stylesheet" href="/styl.css">`),
			},
			staticPaths: []string{"style.css"},
			wantLogged:  []string{"missing-target"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			engine := DefaultEngine()
			engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))
			engine.Strict = test.strict

			err := engine.checkReferences(test.rendered, test.staticPaths)

			if test.wantErr && err == nil {
				t.Errorf("checkReferences() = nil, want an error under strict")
			}
			if !test.wantErr && err != nil {
				t.Errorf("checkReferences() = %v, want nil", err)
			}
			out := buf.String()
			for _, want := range test.wantLogged {
				if !strings.Contains(out, want) {
					t.Errorf("log does not contain %q; got:\n%s", want, out)
				}
			}
			for _, absent := range test.wantAbsent {
				if strings.Contains(out, absent) {
					t.Errorf("log unexpectedly contains %q; got:\n%s", absent, out)
				}
			}
		})
	}
}
