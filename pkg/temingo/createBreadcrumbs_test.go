package temingo

import (
	"reflect"
	"testing"
)

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
				{Name: "a", Path: "/a"},
			},
		},
		{
			name:         "Three levels - two breadcrumbs",
			renderedPath: "a/b/c/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "a", Path: "/a"},
				{Name: "b", Path: "/a/b"},
			},
		},
		{
			name:         "Four levels - three breadcrumbs",
			renderedPath: "a/b/c/d/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "a", Path: "/a"},
				{Name: "b", Path: "/a/b"},
				{Name: "c", Path: "/a/b/c"},
			},
		},
		{
			name:         "Deep nesting - multiple breadcrumbs",
			renderedPath: "blog/posts/2024/january/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "blog", Path: "/blog"},
				{Name: "posts", Path: "/blog/posts"},
				{Name: "2024", Path: "/blog/posts/2024"},
			},
		},
		{
			name:         "Path with underscores and hyphens",
			renderedPath: "my-blog/posts_2024/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "my-blog", Path: "/my-blog"},
			},
		},
		{
			name:         "Path with numbers",
			renderedPath: "section1/section2/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "section1", Path: "/section1"},
			},
		},
		{
			name:         "Single character directories",
			renderedPath: "x/y/z/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "x", Path: "/x"},
				{Name: "y", Path: "/x/y"},
			},
		},
		{
			name:         "Long directory names",
			renderedPath: "very-long-directory-name/another-long-name/index.html",
			expectedBreadcrumbs: []Breadcrumb{
				{Name: "very-long-directory-name", Path: "/very-long-directory-name"},
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
		{Name: "a", Path: "/a"},
		{Name: "b", Path: "/a/b"},
		{Name: "c", Path: "/a/b/c"},
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
