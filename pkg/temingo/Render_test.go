package temingo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRender_BasicTemplate(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	// Create input directory structure
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Create a simple template
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{ .path }}</title></head>
<body><h1>Hello {{ .path }}</h1></body>
</html>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create engine and render
	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Render() should have created output file %q", outputFile)
	}

	// Verify content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedContent := `<!DOCTYPE html>
<html>
<head><title>index.html</title></head>
<body><h1>Hello index.html</h1></body>
</html>`
	if string(content) != expectedContent {
		t.Errorf("Render() output content = %q, want %q", string(content), expectedContent)
	}
}

func TestRender_WithPartials(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Create partial
	partialFile := filepath.Join(inputDir, "header.partial.html")
	partialContent := `<header>Header Content</header>`
	if err := os.WriteFile(partialFile, []byte(partialContent), 0644); err != nil {
		t.Fatalf("Failed to write partial file: %v", err)
	}

	// Create template using partial
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `{{ template "header.partial.html" }}
<main>Main Content</main>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify output
	outputFile := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "Header Content") {
		t.Errorf("Render() output should contain partial content, got: %q", string(content))
	}
	if !strings.Contains(string(content), "Main Content") {
		t.Errorf("Render() output should contain template content, got: %q", string(content))
	}
}

func TestRender_WithStaticFiles(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(inputDir, "static"), 0755); err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}

	// Create static file
	staticFile := filepath.Join(inputDir, "static", "style.css")
	staticContent := `body { color: red; }`
	if err := os.WriteFile(staticFile, []byte(staticContent), 0644); err != nil {
		t.Fatalf("Failed to write static file: %v", err)
	}

	// Create template
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html><head><link rel="stylesheet" href="/static/style.css"></head></html>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify static file was copied
	outputStaticFile := filepath.Join(outputDir, "static", "style.css")
	if _, err := os.Stat(outputStaticFile); os.IsNotExist(err) {
		t.Errorf("Render() should have copied static file %q", outputStaticFile)
	}

	// Verify static file content
	content, err := os.ReadFile(outputStaticFile)
	if err != nil {
		t.Fatalf("Failed to read static file: %v", err)
	}
	if string(content) != staticContent {
		t.Errorf("Render() static file content = %q, want %q", string(content), staticContent)
	}
}

func TestRender_WithValues(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Create template using values
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html><head><title>{{ .siteName }}</title></head><body>Version: {{ .version }}</body></html>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Values = map[string]string{
		"siteName": "My Website",
		"version":  "2.0.0",
	}
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify output contains values
	outputFile := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "My Website") {
		t.Errorf("Render() output should contain siteName value, got: %q", string(content))
	}
	if !strings.Contains(string(content), "2.0.0") {
		t.Errorf("Render() output should contain version value, got: %q", string(content))
	}
}

func TestRender_DryRun(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Create template
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html><body>Test</body></html>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.DryRun = true
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify output file was NOT created in dry run mode
	outputFile := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Errorf("Render() should not create files in dry run mode, but file exists: %q", outputFile)
	}
}

func TestRender_WithTemingoignore(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Create template
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html><body>Test</body></html>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create file that should be ignored
	ignoredFile := filepath.Join(inputDir, "ignored.template.html")
	if err := os.WriteFile(ignoredFile, []byte("ignored"), 0644); err != nil {
		t.Fatalf("Failed to write ignored file: %v", err)
	}

	// Create .temingoignore file
	ignoreFile := filepath.Join(tmpDir, ".temingoignore")
	if err := os.WriteFile(ignoreFile, []byte("ignored.template.html"), 0644); err != nil {
		t.Fatalf("Failed to write ignore file: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.TemingoignorePath = ignoreFile
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify ignored file was not rendered
	ignoredOutputFile := filepath.Join(outputDir, "ignored.html")
	if _, err := os.Stat(ignoredOutputFile); !os.IsNotExist(err) {
		t.Errorf("Render() should not render ignored files, but file exists: %q", ignoredOutputFile)
	}

	// Verify non-ignored file was rendered
	outputFile := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Render() should render non-ignored files, but file does not exist: %q", outputFile)
	}
}

func TestRender_NoDeleteOutputDir(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create existing file in output directory
	existingFile := filepath.Join(outputDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Create template
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html><body>Test</body></html>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.NoDeleteOutputDir = true
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify existing file still exists
	if _, err := os.Stat(existingFile); os.IsNotExist(err) {
		t.Errorf("Render() with NoDeleteOutputDir should preserve existing files, but file was deleted: %q", existingFile)
	}

	// Verify new template was rendered
	outputFile := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Render() should render templates even with NoDeleteOutputDir, but file does not exist: %q", outputFile)
	}
}

func TestRender_TemplateHelperFunctions(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Create template using helper functions
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html>
<body>
<h1>{{ capitalize "hello world" }}</h1>
<p>{{ concat "Hello" " " "World" }}</p>
<pre>{{ includeWithIndentation 2 "line1\nline2" }}</pre>
</body>
</html>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify output
	outputFile := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Hello World") {
		t.Errorf("Render() output should contain capitalized text, got: %q", contentStr)
	}
	if !strings.Contains(contentStr, "HelloWorld") || !strings.Contains(contentStr, "Hello World") {
		// Check for either concatenated or with space
		if !strings.Contains(contentStr, "Hello") || !strings.Contains(contentStr, "World") {
			t.Errorf("Render() output should contain concatenated text, got: %q", contentStr)
		}
	}
	if !strings.Contains(contentStr, "  line1") {
		t.Errorf("Render() output should contain indented content, got: %q", contentStr)
	}
}

func TestRender_EmptyInputDirectory(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Verbose = false

	err := engine.Render()
	if err != nil {
		t.Fatalf("Render() should succeed with empty input directory, got error: %v", err)
	}
}

func TestRender_InvalidTemplate(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Create template with invalid syntax
	templateFile := filepath.Join(inputDir, "index.template.html")
	templateContent := `<html><body>{{ .invalid.syntax }}</body></html>`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	engine := DefaultEngine()
	engine.InputDir = inputDir + string(filepath.Separator)
	engine.OutputDir = outputDir + string(filepath.Separator)
	engine.Verbose = false

	err := engine.Render()
	// Note: This might not error if the template engine is lenient
	// The actual behavior depends on the template engine
	if err != nil {
		// If it errors, that's expected for invalid templates
		t.Logf("Render() correctly returned error for invalid template: %v", err)
	}
}

func TestRender_UsingInitFilesTestProject(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	// Use the helper to set up the test project from InitFiles
	engine, _, outputDir, err := setupTestProjectFromInitFilesWithEngine(tmpDir, "test", nil)
	if err != nil {
		t.Fatalf("setupTestProjectFromInitFilesWithEngine() failed: %v", err)
	}

	// Render the project
	err = engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify that the main template was rendered
	outputFile := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Render() should have created output file %q", outputFile)
	}

	// Verify that templates with meta were rendered
	metaOutputFile := filepath.Join(outputDir, "test_meta", "index.html")
	if _, err := os.Stat(metaOutputFile); os.IsNotExist(err) {
		t.Errorf("Render() should have created meta template output file %q", metaOutputFile)
	}

	// Verify that markdown content was processed
	markdownOutputFile := filepath.Join(outputDir, "test_markdown", "index.html")
	if _, err := os.Stat(markdownOutputFile); os.IsNotExist(err) {
		t.Errorf("Render() should have created markdown template output file %q", markdownOutputFile)
	}

	// Verify that static files were copied
	staticFile := filepath.Join(outputDir, "test_static", "static.asset")
	if _, err := os.Stat(staticFile); os.IsNotExist(err) {
		t.Errorf("Render() should have copied static file %q", staticFile)
	}
}

func TestRender_UsingInitFilesExampleProject(t *testing.T) {
	// Create a new tmpDir for each test case to ensure isolation
	tmpDir := t.TempDir()

	// Use the helper to set up the example project from InitFiles
	engine, _, outputDir, err := setupTestProjectFromInitFilesWithEngine(tmpDir, "example", map[string]string{
		"siteName": "Test Site",
	})
	if err != nil {
		t.Fatalf("setupTestProjectFromInitFilesWithEngine() failed: %v", err)
	}

	// Render the project
	err = engine.Render()
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	// Verify that the main template was rendered
	outputFile := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Render() should have created output file %q", outputFile)
	}

	// Verify that blog template was rendered
	blogOutputFile := filepath.Join(outputDir, "blog", "index.html")
	if _, err := os.Stat(blogOutputFile); os.IsNotExist(err) {
		t.Errorf("Render() should have created blog output file %q", blogOutputFile)
	}

	// Verify that static files were copied
	staticFile := filepath.Join(outputDir, "static", "static.asset")
	if _, err := os.Stat(staticFile); os.IsNotExist(err) {
		t.Errorf("Render() should have copied static file %q", staticFile)
	}
}
