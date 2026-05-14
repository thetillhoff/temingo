package temingo

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestWarnUnusedPartials(t *testing.T) {
	tests := []struct {
		name             string
		partialFiles     map[string]string
		templateContents []string
		wantWarnings     []string
		wantNoWarn       bool
	}{
		{
			name:         "no partials",
			partialFiles: map[string]string{},
			wantNoWarn:   true,
		},
		{
			name: "all partials used",
			partialFiles: map[string]string{
				"header.partial.html": `{{ define "header.partial.html" }}<header></header>{{ end }}`,
			},
			templateContents: []string{`{{ template "header.partial.html" . }}`},
			wantNoWarn:       true,
		},
		{
			name: "unused partial warns",
			partialFiles: map[string]string{
				"header.partial.html": `{{ define "header.partial.html" }}<header></header>{{ end }}`,
			},
			templateContents: []string{`<html><body>no partials here</body></html>`},
			wantWarnings:     []string{"header.partial.html"},
		},
		{
			name: "partial used by another partial not warned",
			partialFiles: map[string]string{
				"header.partial.html": `{{ define "header.partial.html" }}{{ template "nav.partial.html" . }}{{ end }}`,
				"nav.partial.html":    `{{ define "nav.partial.html" }}<nav></nav>{{ end }}`,
			},
			templateContents: []string{`{{ template "header.partial.html" . }}`},
			wantNoWarn:       true,
		},
		{
			name: "multiple unused partials",
			partialFiles: map[string]string{
				"header.partial.html": `{{ define "header.partial.html" }}<header></header>{{ end }}`,
				"footer.partial.html": `{{ define "footer.partial.html" }}<footer></footer>{{ end }}`,
			},
			templateContents: []string{`<html>no partials</html>`},
			wantWarnings:     []string{"header.partial.html", "footer.partial.html"},
		},
		{
			name: "partial used in metatemplate",
			partialFiles: map[string]string{
				"card.partial.html": `{{ define "card.partial.html" }}<div></div>{{ end }}`,
			},
			templateContents: []string{
				`<html>no partials</html>`,
				`{{ template "card.partial.html" . }}`,
			},
			wantNoWarn: true,
		},
		{
			name: "whitespace variants in template call",
			partialFiles: map[string]string{
				"nav.partial.html": `{{ define "nav.partial.html" }}<nav></nav>{{ end }}`,
			},
			templateContents: []string{`{{- template "nav.partial.html" . -}}`},
			wantNoWarn:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

			engine := DefaultEngine()
			engine.Logger = logger

			engine.warnUnusedPartials(tt.partialFiles, tt.templateContents)

			output := buf.String()
			if tt.wantNoWarn {
				if output != "" {
					t.Errorf("warnUnusedPartials() logged unexpected warnings: %q", output)
				}
				return
			}
			for _, want := range tt.wantWarnings {
				if !strings.Contains(output, want) {
					t.Errorf("warnUnusedPartials() log output missing %q, got: %q", want, output)
				}
			}
		})
	}
}
