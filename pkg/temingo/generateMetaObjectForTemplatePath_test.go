package temingo

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/thetillhoff/fileIO"
)

func TestGenerateMetaObjectForTemplatePath(t *testing.T) {
	tests := []struct {
		name                 string
		renderedTemplatePath string
		setup                func(tmpDir string) (fileIO.FileList, []string, error)
		engineValues         map[string]string
		wantPath             string
		wantBreadcrumbs      []Breadcrumb
		wantHasContent       bool
		wantHasMeta          bool
		wantHasChildMeta     bool
		description          string
	}{
		{
			name:                 "Basic template with path only",
			renderedTemplatePath: "index.html",
			setup: func(tmpDir string) (fileIO.FileList, []string, error) {
				// No files needed for basic test
				return fileIO.FileList{Files: []string{}}, []string{}, nil
			},
			engineValues:     map[string]string{},
			wantPath:         "index.html",
			wantBreadcrumbs:  []Breadcrumb{},
			wantHasContent:   false,
			wantHasMeta:      false,
			wantHasChildMeta: false,
			description:      "Basic template should have path and empty breadcrumbs",
		},
		{
			name:                 "Template with breadcrumbs",
			renderedTemplatePath: "blog/posts/index.html",
			setup: func(tmpDir string) (fileIO.FileList, []string, error) {
				return fileIO.FileList{Files: []string{}}, []string{}, nil
			},
			engineValues: map[string]string{},
			wantPath:     "blog/posts/index.html",
			wantBreadcrumbs: []Breadcrumb{
				{Name: "blog", Path: "/blog/"},
			},
			wantHasContent:   false,
			wantHasMeta:      false,
			wantHasChildMeta: false,
			description:      "Template in nested directory should have breadcrumbs",
		},
		{
			name:                 "Template with markdown content",
			renderedTemplatePath: "about/index.html",
			setup: func(tmpDir string) (fileIO.FileList, []string, error) {
				inputDir := filepath.Join(tmpDir, "input")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					return fileIO.FileList{}, []string{}, err
				}
				if err := os.MkdirAll(filepath.Join(inputDir, "about"), 0755); err != nil {
					return fileIO.FileList{}, []string{}, err
				}

				// Create markdown content file
				contentFile := filepath.Join(inputDir, "about", "content.md")
				err := os.WriteFile(contentFile, []byte("# About\n\nThis is about content."), 0644)
				if err != nil {
					return fileIO.FileList{}, []string{}, err
				}

				return fileIO.FileList{
					Files: []string{"about/content.md"},
					Path:  inputDir,
				}, []string{}, nil
			},
			engineValues:     map[string]string{},
			wantPath:         "about/index.html",
			wantBreadcrumbs:  []Breadcrumb{},
			wantHasContent:   true,
			wantHasMeta:      false,
			wantHasChildMeta: false,
			description:      "Template with markdown content should have content field",
		},
		{
			name:                 "Template with meta.yaml",
			renderedTemplatePath: "blog/index.html",
			setup: func(tmpDir string) (fileIO.FileList, []string, error) {
				inputDir := filepath.Join(tmpDir, "input")
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					return fileIO.FileList{}, []string{}, err
				}
				if err := os.MkdirAll(filepath.Join(inputDir, "blog"), 0755); err != nil {
					return fileIO.FileList{}, []string{}, err
				}

				// Create meta.yaml file
				metaFile := filepath.Join(inputDir, "blog", "meta.yaml")
				err := os.WriteFile(metaFile, []byte("title: Blog\nauthor: John"), 0644)
				if err != nil {
					return fileIO.FileList{}, []string{}, err
				}

				return fileIO.FileList{
					Files: []string{},
					Path:  inputDir,
				}, []string{"blog/meta.yaml"}, nil
			},
			engineValues:     map[string]string{},
			wantPath:         "blog/index.html",
			wantBreadcrumbs:  []Breadcrumb{},
			wantHasContent:   false,
			wantHasMeta:      true,
			wantHasChildMeta: false,
			description:      "Template with meta.yaml should have meta field",
		},
		{
			name:                 "Template with values",
			renderedTemplatePath: "index.html",
			setup: func(tmpDir string) (fileIO.FileList, []string, error) {
				return fileIO.FileList{Files: []string{}}, []string{}, nil
			},
			engineValues: map[string]string{
				"siteName": "My Site",
				"version":  "1.0.0",
			},
			wantPath:         "index.html",
			wantBreadcrumbs:  []Breadcrumb{},
			wantHasContent:   false,
			wantHasMeta:      false,
			wantHasChildMeta: false,
			description:      "Template with engine values should include them in meta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new tmpDir for each test case to ensure isolation
			tmpDir := t.TempDir()

			fileList, metaPaths, err := tt.setup(tmpDir)
			if err != nil {
				t.Fatalf("Test setup failed: %v", err)
			}

			engine := DefaultEngine()
			engine.InputDir = filepath.Join(tmpDir, "input") + string(filepath.Separator)
			engine.Values = tt.engineValues

			meta, err := engine.generateMetaObjectForTemplatePath(tt.renderedTemplatePath, fileList, metaPaths)
			if err != nil {
				t.Fatalf("generateMetaObjectForTemplatePath() unexpected error: %v", err)
			}

			// Check path
			if meta["path"] != tt.wantPath {
				t.Errorf("meta[\"path\"] = %q, want %q", meta["path"], tt.wantPath)
			}

			// Check breadcrumbs
			breadcrumbs, ok := meta["breadcrumbs"].([]Breadcrumb)
			if !ok {
				t.Errorf("meta[\"breadcrumbs\"] is not []Breadcrumb")
			} else {
				if !reflect.DeepEqual(breadcrumbs, tt.wantBreadcrumbs) {
					t.Errorf("meta[\"breadcrumbs\"] = %v, want %v", breadcrumbs, tt.wantBreadcrumbs)
				}
			}

			// Check content
			hasContent := meta["content"] != nil
			if hasContent != tt.wantHasContent {
				t.Errorf("meta has content = %v, want %v", hasContent, tt.wantHasContent)
			}

			// Check meta
			hasMeta := meta["meta"] != nil
			if hasMeta != tt.wantHasMeta {
				t.Errorf("meta has meta = %v, want %v", hasMeta, tt.wantHasMeta)
			}

			// Check childMeta (it's always present, but may be empty)
			childMeta, ok := meta["childMeta"].(map[string]interface{})
			if !ok {
				t.Errorf("meta[\"childMeta\"] is not map[string]interface{}")
			} else {
				hasChildMeta := len(childMeta) > 0
				if hasChildMeta != tt.wantHasChildMeta {
					t.Errorf("meta has non-empty childMeta = %v, want %v", hasChildMeta, tt.wantHasChildMeta)
				}
			}

			// Check values
			for key, expectedValue := range tt.engineValues {
				if meta[key] != expectedValue {
					t.Errorf("meta[%q] = %q, want %q", key, meta[key], expectedValue)
				}
			}
		})
	}
}

func TestCreateBreadcrumbs(t *testing.T) {
	tests := []struct {
		name                string
		renderedPath        string
		expectedBreadcrumbs []Breadcrumb
	}{
		{
			name:                "Root index.html - empty breadcrumbs",
			renderedPath:        "index.html",
			expectedBreadcrumbs: []Breadcrumb{},
		},
		{
			name:                "Single level - empty breadcrumbs",
			renderedPath:        "a/index.html",
			expectedBreadcrumbs: []Breadcrumb{},
		},
		{
			name:         "Two levels - one breadcrumb",
			renderedPath: "a/b/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "a", Path: "/a/"},
			},
		},
		{
			name:         "Three levels - two breadcrumbs",
			renderedPath: "a/b/c/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "a", Path: "/a/"},
				{Name: "b", Path: "/a/b/"},
			},
		},
		{
			name:         "Four levels - three breadcrumbs",
			renderedPath: "a/b/c/d/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "a", Path: "/a/"},
				{Name: "b", Path: "/a/b/"},
				{Name: "c", Path: "/a/b/c/"},
			},
		},
		{
			name:         "Deep nesting - multiple breadcrumbs",
			renderedPath: "blog/posts/2024/january/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "blog", Path: "/blog/"},
				{Name: "posts", Path: "/blog/posts/"},
				{Name: "2024", Path: "/blog/posts/2024/"},
			},
		},
		{
			name:         "Path with underscores and hyphens",
			renderedPath: "my-blog/posts_2024/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "my-blog", Path: "/my-blog/"},
			},
		},
		{
			name:         "Path with numbers",
			renderedPath: "section1/section2/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "section1", Path: "/section1/"},
			},
		},
		{
			name:         "Single character directories",
			renderedPath: "x/y/z/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "x", Path: "/x/"},
				{Name: "y", Path: "/x/y/"},
			},
		},
		{
			name:         "Long directory names",
			renderedPath: "very-long-directory-name/another-long-name/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "very-long-directory-name", Path: "/very-long-directory-name/"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := createBreadcrumbs(test.renderedPath)

			if !reflect.DeepEqual(result, test.expectedBreadcrumbs) {
				t.Errorf("createBreadcrumbs(%q) = %v, want %v", test.renderedPath, result, test.expectedBreadcrumbs)
			}
		})
	}
}

func TestCreateBreadcrumbsPathStructure(t *testing.T) {
	// Test that paths are built incrementally and correctly
	result := createBreadcrumbs("a/b/c/d/index.html")

	if len(result) != 3 {
		t.Fatalf("Expected 3 breadcrumbs, got %d", len(result))
	}

	expected := []Breadcrumb{
		{Name: "a", Path: "/a/"},
		{Name: "b", Path: "/a/b/"},
		{Name: "c", Path: "/a/b/c/"},
	}

	for i, breadcrumb := range result {
		if breadcrumb.Name != expected[i].Name {
			t.Errorf("Breadcrumb[%d].Name = %q, want %q", i, breadcrumb.Name, expected[i].Name)
		}
		if breadcrumb.Path != expected[i].Path {
			t.Errorf("Breadcrumb[%d].Path = %q, want %q", i, breadcrumb.Path, expected[i].Path)
		}
	}
}

func TestCreateBreadcrumbsEmptyCases(t *testing.T) {
	emptyCases := []string{
		"index.html",
		"a/index.html",
		"./index.html",
	}

	for _, path := range emptyCases {
		t.Run(path, func(t *testing.T) {
			result := createBreadcrumbs(path)
			if len(result) != 0 {
				t.Errorf("createBreadcrumbs(%q) = %v, want empty slice", path, result)
			}
		})
	}
}
