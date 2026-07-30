package temingo

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

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
