package temingo

import (
	"testing"

	"github.com/thetillhoff/fileIO/v2"
)

// Check if sorting works if all types of paths are passed
func TestSortPathWithAll(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{
			"src/index.template.html",
			"src/index.metatemplate.html",
			"src/partials/some.partial.css",
			"src/meta.yaml",
			"src/static.asset",
		},
	}

	templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 1 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 1, got", len(templatePaths))
		if templatePaths[0] != "src/index.template.html" {
			t.Fatal("wrong return value of sortPath templatePaths: expected src/index.template.html, got", templatePaths[0])
		}
	} else if len(metaTemplatePaths) != 1 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 1, got", len(metaTemplatePaths))
		if metaTemplatePaths[0] != "src/index.metatemplate.html" {
			t.Fatal("wrong return value of sortPath metaTemplatePaths: expected src/index.metatemplate.html, got", metaTemplatePaths[0])
		}
	} else if len(partialPaths) != 1 {
		t.Fatal("wrong amount of partialPaths returned from sortPaths: expected 1, got", len(partialPaths))
		if partialPaths[0] != "src/partials/some.partial.css" {
			t.Fatal("wrong return value of sortPath partialPaths: expected src/partials/some.partial.css, got", partialPaths[0])
		}
	} else if len(metaPaths) != 1 {
		t.Fatal("wrong amount of metaPaths returned from sortPaths: expected 1, got", len(metaPaths))
		if metaPaths[0] != "" {
			t.Fatal("wrong return value of sortPath metaPaths: expected src/meta.yaml, got", staticPaths[0])
		}
	} else if len(staticPaths) != 1 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 1, got", len(staticPaths))
		if staticPaths[0] != "src/static.asset" {
			t.Fatal("wrong return value of sortPath staticPaths: expected src/static.asset, got", staticPaths[0])
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

	templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 1 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 1, got", len(templatePaths))
		if templatePaths[0] != "src/index.template.html" {
			t.Fatal("wrong return value of sortPaths: expected src/index.template.html, got", templatePaths[0])
		}
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(partialPaths) != 0 {
		t.Fatal("wrong amount of partialPaths returned from sortPaths: expected 0, got", len(partialPaths))
	} else if len(metaPaths) != 0 {
		t.Fatal("wrong amount of metaPaths returned from sortPaths: expected 0, got", len(partialPaths))
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

	templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 1 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 1, got", len(metaTemplatePaths))
		if metaTemplatePaths[0] != "src/index.metatemplate.html" {
			t.Fatal("wrong return value of sortPaths: expected src/index.metatemplate.html, got", metaTemplatePaths[0])
		}
	} else if len(partialPaths) != 0 {
		t.Fatal("wrong amount of partialPaths returned from sortPaths: expected 0, got", len(partialPaths))
	} else if len(metaPaths) != 0 {
		t.Fatal("wrong amount of metaPaths returned from sortPaths: expected 0, got", len(partialPaths))
	} else if len(staticPaths) != 0 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 0, got", len(staticPaths))
	}

}

// Check if sorting works if only one type of path is passed - partials
func TestSortPathWithOnlyPartials(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{
			"src/partials/some.partial.css",
		},
	}

	templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(partialPaths) != 1 {
		t.Fatal("wrong amount of partialPaths returned from sortPaths: expected 1, got", len(partialPaths))
		if partialPaths[0] != "src/partials/some.partial.css" {
			t.Fatal("wrong return value of sortPaths: expected src/partials/some.partial.css, got", partialPaths[0])
		}
	} else if len(metaPaths) != 0 {
		t.Fatal("wrong amount of metaPaths returned from sortPaths: expected 0, got", len(partialPaths))
	} else if len(staticPaths) != 0 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 0, got", len(staticPaths))
	}

}

// Check if sorting works if only one type of path is passed - meta
func TestSortPathWithOnlyMeta(t *testing.T) {

	engine := DefaultEngine()

	fileList := fileIO.FileList{
		Files: []string{
			"src/meta.yaml",
		},
	}

	templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(partialPaths) != 0 {
		t.Fatal("wrong amount of partialPaths returned from sortPaths: expected 0, got", len(partialPaths))
	} else if len(metaPaths) != 1 {
		t.Fatal("wrong amount of metaPaths returned from sortPaths: expected 1, got", len(partialPaths))
		if metaPaths[0] != "src/meta.yaml" {
			t.Fatal("wrong return value of sortPaths: expected src/meta.yaml, got", metaPaths[0])
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

	templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(partialPaths) != 0 {
		t.Fatal("wrong amount of partialPaths returned from sortPaths: expected 0, got", len(partialPaths))
	} else if len(metaPaths) != 0 {
		t.Fatal("wrong amount of metaPaths returned from sortPaths: expected 0, got", len(partialPaths))
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

	templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths := engine.sortPaths(fileList)

	if len(templatePaths) != 0 {
		t.Fatal("wrong amount of templatePaths returned from sortPaths: expected 0, got", len(templatePaths))
	} else if len(metaTemplatePaths) != 0 {
		t.Fatal("wrong amount of metaTemplatePaths returned from sortPaths: expected 0, got", len(metaTemplatePaths))
	} else if len(partialPaths) != 0 {
		t.Fatal("wrong amount of partialPaths returned from sortPaths: expected 0, got", len(partialPaths))
	} else if len(metaPaths) != 0 {
		t.Fatal("wrong amount of metaPaths returned from sortPaths: expected 0, got", len(metaPaths))
	} else if len(staticPaths) != 0 {
		t.Fatal("wrong amount of staticPaths returned from sortPaths: expected 0, got", len(staticPaths))
	}

}
