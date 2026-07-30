package prettifyhtml

import (
	"strings"
	"testing"
)

func TestFormatPreservesTextEntities(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		wants []string
	}{
		{
			name:  "escaped markup in prose stays text",
			in:    "<p>&lt;div&gt;</p>",
			wants: []string{"&lt;div&gt;"},
		},
		{
			name:  "ampersand round-trips",
			in:    "<p>&amp;amp;</p>",
			wants: []string{"&amp;amp;"},
		},
		{
			name:  "child selector in style stays valid",
			in:    "<style>.a > .b { color: red; }</style>",
			wants: []string{".a > .b"},
		},
		{
			name:  "script operators stay valid",
			in:    "<script>if (a > b && c < d) {}</script>",
			wants: []string{"a > b && c < d"},
		},
		{
			name:  "pre content keeps exact whitespace",
			in:    "<pre>a\n  b</pre>",
			wants: []string{"<pre>a\n  b</pre>"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := Format(tt.in)
			for _, want := range tt.wants {
				if !strings.Contains(out, want) {
					t.Errorf("Format(%q) = %q, want it to contain %q", tt.in, out, want)
				}
			}
		})
	}
}
