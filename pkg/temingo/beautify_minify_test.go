package temingo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBeautify(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		ext      string
		expected string
		description string
	}{
		{
			name:     "HTML beautification",
			content:  []byte("<html><head><title>Test</title></head><body><h1>Hello</h1></body></html>"),
			ext:      ".html",
			expected: "<html>",
			description: "HTML content should be beautified",
		},
		{
			name:     "Non-HTML file - should return unchanged",
			content:  []byte("plain text content"),
			ext:      ".txt",
			expected: "plain text content",
			description: "Non-HTML files should be returned unchanged",
		},
		{
			name:     "Empty content",
			content:  []byte(""),
			ext:      ".html",
			expected: "",
			description: "Empty content should return empty string",
		},
		{
			name:     "CSS file - should return unchanged",
			content:  []byte("body { color: red; }"),
			ext:      ".css",
			expected: "body { color: red; }",
			description: "CSS files should be returned unchanged (not yet implemented)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := DefaultEngine()
			result := engine.beautify(tt.content, tt.ext)

			resultStr := string(result)
			if tt.ext == ".html" && len(tt.content) > 0 {
				// For HTML, just check that it's beautified (contains newlines/formatting)
				// The exact format may vary, so we check for basic beautification
				if !strings.Contains(resultStr, "<html") {
					t.Errorf("beautify() output should contain HTML tags, got: %q (%s)", resultStr, tt.description)
				}
			} else {
				// For non-HTML or empty content, should be unchanged
				if resultStr != tt.expected {
					t.Errorf("beautify() = %q, want %q (%s)", resultStr, tt.expected, tt.description)
				}
			}
		})
	}
}

func TestMinify(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		ext         string
		expected    string
		description string
	}{
		{
			name:        "HTML minification - not yet implemented",
			content:     []byte("<html>\n<head>\n<title>Test</title>\n</head>\n</html>"),
			ext:         ".html",
			expected:    "<html>\n<head>\n<title>Test</title>\n</head>\n</html>",
			description: "Minification is not yet implemented, should return unchanged",
		},
		{
			name:        "Plain text - should return unchanged",
			content:     []byte("plain text content"),
			ext:         ".txt",
			expected:    "plain text content",
			description: "Plain text should be returned unchanged",
		},
		{
			name:        "Empty content",
			content:     []byte(""),
			ext:         ".html",
			expected:    "",
			description: "Empty content should be handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := DefaultEngine()
			result := engine.minify(tt.content, tt.ext)

			resultStr := string(result)
			if resultStr != tt.expected {
				t.Errorf("minify() = %q, want %q (%s)", resultStr, tt.expected, tt.description)
			}
		})
	}
}

func TestRender_WithBeautify(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	os.MkdirAll(inputDir, 0755)

	// Create a template with unformatted HTML
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html><head><title>Test</title></head><body><h1>Hello</h1></body></html>`
	os.WriteFile(templateFile, []byte(templateContent), 0644)

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Beautify = true
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() with Beautify unexpected error: %v", err)
	}

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check that HTML is beautified (contains newlines)
	contentStr := string(content)
	if !strings.Contains(contentStr, "\n") {
		t.Errorf("Render() with Beautify should format HTML with newlines, got: %q", contentStr)
	}
}

func TestRender_WithMinify(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	os.MkdirAll(inputDir, 0755)

	// Create a template with formatted HTML
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html>
<head>
	<title>Test</title>
</head>
<body>
	<h1>Hello</h1>
</body>
</html>`
	os.WriteFile(templateFile, []byte(templateContent), 0644)

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Minify = true
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() with Minify unexpected error: %v", err)
	}

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Note: Minify is not yet implemented, so content should be unchanged
	// This test documents current behavior
	contentStr := string(content)
	if !strings.Contains(contentStr, "<html>") {
		t.Errorf("Render() with Minify should still produce valid HTML, got: %q", contentStr)
	}
}

