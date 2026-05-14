package temingo

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestWarnHTTPLinks(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantWarnings []string // substrings expected in log output
		wantNoWarn   bool     // expect no warnings at all
	}{
		{
			name:       "no links",
			content:    `<html><body><p>No links here</p></body></html>`,
			wantNoWarn: true,
		},
		{
			name:       "https link only",
			content:    `<a href="https://example.com">link</a>`,
			wantNoWarn: true,
		},
		{
			name:         "insecure http link",
			content:      `<a href="http://example.com/page">link</a>`,
			wantWarnings: []string{"http://example.com/page"},
		},
		{
			name:       "localhost http excluded",
			content:    `<a href="http://localhost:3000">local</a>`,
			wantNoWarn: true,
		},
		{
			name:       "loopback http excluded",
			content:    `<a href="http://127.0.0.1:8080">local</a>`,
			wantNoWarn: true,
		},
		{
			name:    "multiple insecure links",
			content: `<a href="http://foo.com">a</a> <a href="http://bar.com">b</a>`,
			wantWarnings: []string{
				"http://foo.com",
				"http://bar.com",
			},
		},
		{
			name:         "mixed secure and insecure with localhost",
			content:      `<a href="https://safe.com">s</a><a href="http://localhost">l</a><a href="http://unsafe.com">u</a>`,
			wantWarnings: []string{"http://unsafe.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

			engine := DefaultEngine()
			engine.Logger = logger

			engine.warnHTTPLinks("test.html", []byte(tt.content))

			output := buf.String()
			if tt.wantNoWarn {
				if output != "" {
					t.Errorf("warnHTTPLinks() logged unexpected warnings: %q", output)
				}
				return
			}
			for _, want := range tt.wantWarnings {
				if !strings.Contains(output, want) {
					t.Errorf("warnHTTPLinks() log output missing %q, got: %q", want, output)
				}
			}
		})
	}
}
