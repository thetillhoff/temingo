package temingo

import (
	"testing"
)

func TestTmplConcat(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "Empty arguments",
			args:     []string{},
			expected: "",
		},
		{
			name:     "Single argument",
			args:     []string{"hello"},
			expected: "hello",
		},
		{
			name:     "Two arguments",
			args:     []string{"hello", "world"},
			expected: "helloworld",
		},
		{
			name:     "Multiple arguments",
			args:     []string{"a", "b", "c", "d"},
			expected: "abcd",
		},
		{
			name:     "Empty strings",
			args:     []string{"", "", ""},
			expected: "",
		},
		{
			name:     "Mixed empty and non-empty",
			args:     []string{"a", "", "b", ""},
			expected: "ab",
		},
		{
			name:     "Numbers as strings",
			args:     []string{"1", "2", "3"},
			expected: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tmpl_concat(tt.args...)
			if result != tt.expected {
				t.Errorf("tmpl_concat() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTmplIndent(t *testing.T) {
	tests := []struct {
		name        string
		indentation int
		content     string
		expected    string
	}{
		{
			name:        "Zero indentation",
			indentation: 0,
			content:     "line1\nline2\nline3",
			expected:    "line1\nline2\nline3",
		},
		{
			name:        "Single space indentation",
			indentation: 1,
			content:     "line1\nline2",
			expected:    " line1\n line2",
		},
		{
			name:        "Four space indentation",
			indentation: 4,
			content:     "line1\nline2\nline3",
			expected:    "    line1\n    line2\n    line3",
		},
		{
			name:        "Empty content",
			indentation: 2,
			content:     "",
			expected:    "  ", // Empty string with indentation returns spaces
		},
		{
			name:        "Single line",
			indentation: 2,
			content:     "single line",
			expected:    "  single line",
		},
		{
			name:        "Content with empty lines",
			indentation: 2,
			content:     "line1\n\nline2",
			expected:    "  line1\n  \n  line2",
		},
		{
			name:        "Content ending with newline",
			indentation: 2,
			content:     "line1\nline2\n",
			expected:    "  line1\n  line2\n  ",
		},
		{
			name:        "Large indentation",
			indentation: 10,
			content:     "line1\nline2",
			expected:    "          line1\n          line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tmpl_indent(tt.indentation, tt.content)
			if result != tt.expected {
				t.Errorf("tmpl_indent(%d, %q) = %q, want %q", tt.indentation, tt.content, result, tt.expected)
			}
		})
	}
}

func TestTmplCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase word",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "Uppercase word",
			input:    "HELLO",
			expected: "Hello",
		},
		{
			name:     "Mixed case",
			input:    "hElLo",
			expected: "Hello",
		},
		{
			name:     "Multiple words",
			input:    "hello world",
			expected: "Hello World",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "Numbers and letters",
			input:    "hello123 world",
			expected: "Hello123 World",
		},
		{
			name:     "Special characters",
			input:    "hello-world",
			expected: "Hello-World",
		},
		{
			name:     "Already capitalized",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Sentence",
			input:    "the quick brown fox",
			expected: "The Quick Brown Fox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tmpl_capitalize(tt.input)
			if result != tt.expected {
				t.Errorf("tmpl_capitalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

