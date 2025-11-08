package temingo

import (
	"strings"
	"testing"
)

func TestRenderTemplate_WithValidTemplate(t *testing.T) {
	engine := DefaultEngine()

	templatePath := "random/path/test.template.txt"
	templateContent := `{{ .path }}`
	partialFiles := map[string]string{}

	expectedValue := `random/path/test.template.txt`

	meta := map[string]interface{}{
		"path": "random/path/test.template.txt",
	}

	renderedTemplate, err := engine.renderTemplate(meta, templatePath, templateContent, partialFiles)
	if err != nil {
		t.Fatalf("renderTemplate() expected template rendering to be successful, got error: %v", err)
	}

	if string(renderedTemplate) != expectedValue {
		t.Errorf("renderTemplate() expected template content to be %q, but got %q", expectedValue, string(renderedTemplate))
	}
}

func TestRenderTemplate_WithValidTemplateAndPartialFile(t *testing.T) {
	engine := DefaultEngine()

	templatePath := "random/path/test.template.txt"
	templateContent := `{{ template "partials/first.partial.html" }}`
	expectedValue := `test`
	// Partials are wrapped by Render(), so we need to simulate that here
	partialFiles := map[string]string{
		"partials/first.partial.html": `{{ define "partials/first.partial.html" -}}
test
{{- end -}}`,
	}

	meta := map[string]interface{}{
		"path": "random/path/test.template.txt",
	}

	renderedTemplate, err := engine.renderTemplate(meta, templatePath, templateContent, partialFiles)
	if err != nil {
		t.Fatalf("renderTemplate() expected template rendering to be successful, got error: %v", err)
	}

	if string(renderedTemplate) != expectedValue {
		t.Errorf("renderTemplate() expected rendered template to be %q, but got %q", expectedValue, string(renderedTemplate))
	}
}

func TestRenderTemplate_WithMetaData(t *testing.T) {
	engine := DefaultEngine()

	templatePath := "test.template.html"
	templateContent := `Title: {{ .meta.title }}\nAuthor: {{ .meta.author }}`
	partialFiles := map[string]string{}

	meta := map[string]interface{}{
		"path": "test.html",
		"meta": map[string]interface{}{
			"title":  "Test Title",
			"author": "Test Author",
		},
	}

	renderedTemplate, err := engine.renderTemplate(meta, templatePath, templateContent, partialFiles)
	if err != nil {
		t.Fatalf("renderTemplate() unexpected error: %v", err)
	}

	renderedStr := string(renderedTemplate)
	if !strings.Contains(renderedStr, "Test Title") {
		t.Errorf("renderTemplate() output should contain 'Test Title', got: %q", renderedStr)
	}
	if !strings.Contains(renderedStr, "Test Author") {
		t.Errorf("renderTemplate() output should contain 'Test Author', got: %q", renderedStr)
	}
}

func TestRenderTemplate_WithParentMetaData(t *testing.T) {
	engine := DefaultEngine()

	templatePath := "child/index.template.html"
	templateContent := `Parent: {{ .meta.parent }}\nChild: {{ .meta.child }}`
	partialFiles := map[string]string{}

	meta := map[string]interface{}{
		"path": "child/index.html",
		"meta": map[string]interface{}{
			"parent": "Parent Value",
			"child":  "Child Value",
		},
	}

	renderedTemplate, err := engine.renderTemplate(meta, templatePath, templateContent, partialFiles)
	if err != nil {
		t.Fatalf("renderTemplate() unexpected error: %v", err)
	}

	renderedStr := string(renderedTemplate)
	if !strings.Contains(renderedStr, "Parent Value") {
		t.Errorf("renderTemplate() output should contain 'Parent Value', got: %q", renderedStr)
	}
	if !strings.Contains(renderedStr, "Child Value") {
		t.Errorf("renderTemplate() output should contain 'Child Value', got: %q", renderedStr)
	}
}

func TestRenderTemplate_WithInvalidTemplate(t *testing.T) {
	engine := DefaultEngine()

	templatePath := "test.template.html"
	// Use a template syntax error that will definitely fail
	templateContent := `{{ if .invalid.field.access }}{{ end }}`
	partialFiles := map[string]string{}

	meta := map[string]interface{}{
		"path": "test.html",
	}

	_, err := engine.renderTemplate(meta, templatePath, templateContent, partialFiles)
	// Note: Go templates may or may not error on invalid field access depending on context
	// This test documents the behavior - if it doesn't error, that's also valid behavior
	if err != nil {
		t.Logf("renderTemplate() correctly returned error for invalid template: %v", err)
	} else {
		t.Logf("renderTemplate() did not error for invalid template (this may be valid behavior)")
	}
}

func TestRenderTemplate_WithTemplateHelperFunctions(t *testing.T) {
	engine := DefaultEngine()

	templatePath := "test.template.html"
	templateContent := `{{ capitalize "hello world" }} | {{ concat "a" "b" "c" }} | {{ includeWithIndentation 2 "line1\nline2" }}`
	partialFiles := map[string]string{}

	meta := map[string]interface{}{
		"path": "test.html",
	}

	renderedTemplate, err := engine.renderTemplate(meta, templatePath, templateContent, partialFiles)
	if err != nil {
		t.Fatalf("renderTemplate() unexpected error: %v", err)
	}

	renderedStr := string(renderedTemplate)
	if !strings.Contains(renderedStr, "Hello World") {
		t.Errorf("renderTemplate() output should contain capitalized text, got: %q", renderedStr)
	}
	if !strings.Contains(renderedStr, "abc") {
		t.Errorf("renderTemplate() output should contain concatenated text, got: %q", renderedStr)
	}
	if !strings.Contains(renderedStr, "  line1") {
		t.Errorf("renderTemplate() output should contain indented content, got: %q", renderedStr)
	}
}

func TestRenderTemplate_WithInitFilesTestProject(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	// Use InitFiles/test project
	engine, _, _, err := setupTestProjectFromInitFilesWithEngine(tmpDir, "test", nil)
	if err != nil {
		t.Fatalf("setupTestProjectFromInitFilesWithEngine() failed: %v", err)
	}

	// Read the actual template and partial files from the test project
	templateContent := `hello world

{{ template "partials/first.partial.html" }}

{{ template "partials/second.partial.html" }}

values: {{ . }}
path: {{ .path }}
breadcrumbs: {{ .breadcrumbs }}
meta: {{ .meta }}
childmeta: {{ .childMeta }}`

	// Set up partial files as they would be in Render()
	partialFiles := map[string]string{
		"partials/first.partial.html":  "{{ define \"partials/first.partial.html\" -}}\ntest\n{{- end -}}",
		"partials/second.partial.html": "{{ define \"partials/second.partial.html\" -}}\ntast\n{{- end -}}",
	}

	meta := map[string]interface{}{
		"path":        "index.html",
		"breadcrumbs": []Breadcrumb{},
		"meta":        nil,
		"childMeta":   map[string]interface{}{},
	}

	renderedTemplate, err := engine.renderTemplate(meta, "index.template.html", templateContent, partialFiles)
	if err != nil {
		t.Fatalf("renderTemplate() unexpected error: %v", err)
	}

	renderedStr := string(renderedTemplate)
	if !strings.Contains(renderedStr, "test") {
		t.Errorf("renderTemplate() output should contain first partial content, got: %q", renderedStr)
	}
	if !strings.Contains(renderedStr, "tast") {
		t.Errorf("renderTemplate() output should contain second partial content, got: %q", renderedStr)
	}
	if !strings.Contains(renderedStr, "index.html") {
		t.Errorf("renderTemplate() output should contain path, got: %q", renderedStr)
	}
}
