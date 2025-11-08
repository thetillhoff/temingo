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
				{Name: "blog", Path: "/blog"},
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
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(filepath.Join(inputDir, "about"), 0755)

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
				os.MkdirAll(inputDir, 0755)
				os.MkdirAll(filepath.Join(inputDir, "blog"), 0755)

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
