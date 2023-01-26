package temingo

import (
	"testing"

	"github.com/thetillhoff/temingo/pkg/fileIO"
)

// Check if sorting works if all types of paths are passed
func TestSortPathWithAll(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{
			"src/index.template.html",
			"src/index.metatemplate.html",
			"src/components/some.component.css",
			"src/static.asset",
		},
	}

	templatePaths, metaTemplatePaths, componentPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 1 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 1, got", len(templatePaths))
		if templatePaths[0] != "src/index.template.html" {
			t.Fatal("wrong return value of sortPaths: expected src/index.template.html, got", templatePaths[0])
		}
	} else if len(metaTemplatePaths) != 1 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 1, got", len(metaTemplatePaths))
		if metaTemplatePaths[0] != "src/index.metatemplate.html" {
			t.Fatal("wrong return value of sortPaths: expected src/index.metatemplate.html, got", metaTemplatePaths[0])
		}
	} else if len(componentPaths) != 1 {
		t.Fatal("wrong amount of componentPaths returned from sortPaths: expected 1, got", len(componentPaths))
		if componentPaths[0] != "src/components/some.component.css" {
			t.Fatal("wrong return value of sortPaths: expected src/components/some.component.css, got", componentPaths[0])
		}
	} else if len(staticPaths) != 1 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 1, got", len(staticPaths))
		if staticPaths[0] != "src/static.asset" {
			t.Fatal("wrong return value of sortPaths: expected src/static.asset, got", staticPaths[0])
		}
	}
}

// Check if sorting works if only one type of path is passed - template
func TestSortPathWithOnlyTemplate(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{
			"src/index.template.html",
		},
	}

	templatePaths, metaTemplatePaths, componentPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 1 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 1, got", len(templatePaths))
		if templatePaths[0] != "src/index.template.html" {
			t.Fatal("wrong return value of sortPaths: expected src/index.template.html, got", templatePaths[0])
		}
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(componentPaths) != 0 {
		t.Fatal("wrong amount of componentPaths returned from sortPaths: expected 0, got", len(componentPaths))
	} else if len(staticPaths) != 0 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 0, got", len(staticPaths))
	}

}

// Check if sorting works if only one type of path is passed - metatemplate
func TestSortPathWithOnlyMetaTemplate(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{
			"src/index.metatemplate.html",
		},
	}

	templatePaths, metaTemplatePaths, componentPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 1 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 1, got", len(metaTemplatePaths))
		if metaTemplatePaths[0] != "src/index.metatemplate.html" {
			t.Fatal("wrong return value of sortPaths: expected src/index.metatemplate.html, got", metaTemplatePaths[0])
		}
	} else if len(componentPaths) != 0 {
		t.Fatal("wrong amount of componentPaths returned from sortPaths: expected 0, got", len(componentPaths))
	} else if len(staticPaths) != 0 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 0, got", len(staticPaths))
	}

}

// Check if sorting works if only one type of path is passed - components
func TestSortPathWithOnlyComponents(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{
			"src/components/some.component.css",
		},
	}

	templatePaths, metaTemplatePaths, componentPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(componentPaths) != 1 {
		t.Fatal("wrong amount of componentPaths returned from sortPaths: expected 1, got", len(componentPaths))
		if componentPaths[0] != "src/components/some.component.css" {
			t.Fatal("wrong return value of sortPaths: expected src/components/some.component.css, got", componentPaths[0])
		}
	} else if len(staticPaths) != 0 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 0, got", len(staticPaths))
	}

}

// Check if sorting works if only one type of path is passed - static
func TestSortPathWithOnlyStatic(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{
			"src/static.asset",
		},
	}

	templatePaths, metaTemplatePaths, componentPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(componentPaths) != 0 {
		t.Fatal("wrong amount of componentPaths returned from sortPaths: expected 0, got", len(componentPaths))
	} else if len(staticPaths) != 1 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 1, got", len(staticPaths))
		if staticPaths[0] != "src/static.asset" {
			t.Fatal("wrong return value of sortPaths: expected src/static.asset, got", staticPaths[0])
		}
	}

}

// Check if sorting works if no path is passed
func TestSortPathWithEmpty(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{},
	}

	templatePaths, metaTemplatePaths, componentPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(componentPaths) != 0 {
		t.Fatal("wrong amount of componentPaths returned from sortPaths: expected 0, got", len(componentPaths))
	} else if len(staticPaths) != 0 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 0, got", len(staticPaths))
	}

}
