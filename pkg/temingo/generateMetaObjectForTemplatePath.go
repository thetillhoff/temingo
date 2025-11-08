package temingo

import (
	"log"
	"path"
	"strings"

	"github.com/thetillhoff/fileIO"
	"github.com/thetillhoff/temingo/pkg/markdown2html"
)

// Breadcrumb represents a single breadcrumb with its name and path
type Breadcrumb struct {
	Name string
	Path string
}

func (engine Engine) generateMetaObjectForTemplatePath(templatePath string, renderedTemplatePath string, fileList fileIO.FileList, metaPaths []string) (map[string]interface{}, error) {
	var (
		err error

		meta map[string]interface{}
	)

	// Create meta values object
	meta = map[string]interface{}{}

	// with .path
	meta["path"] = renderedTemplatePath // Path to the current file (without `src/` or `output/`)

	// with .breadcrumbs
	// example for renderedTemplatePath: a/b/c/index.html
	// expected breadcrumbs in that case: [{Name: "a", Path: "/a"}, {Name: "b", Path: "/a/b"}]
	// c should not be added, as the index.html is meant for that folder
	meta["breadcrumbs"] = createBreadcrumbs(renderedTemplatePath)

	// with .meta and .childMeta
	if engine.Verbose {
		log.Println("Searching metadata for", renderedTemplatePath)
	}
	meta["meta"], meta["childMeta"], err = engine.getMetaForTemplatePath(fileIO.FileList{Files: metaPaths}, renderedTemplatePath) // Contains aggregated meta yamls (up to parent dir, were children overwrite their parents values during the merge)
	if err != nil {
		return meta, err
	}

	// with .content
	markdownContentFiles := fileList.FilterByFolderPath(path.Dir(renderedTemplatePath)).FilterByFilename(engine.MarkdownContentFilename).Files
	if len(markdownContentFiles) == 1 { // Can only be 1 at max
		if engine.Verbose {
			log.Println("Getting markdown content for", renderedTemplatePath)
		}
		markdownContent, err := fileIO.ReadFile(path.Join(engine.InputDir, markdownContentFiles[0])) // Read file contents
		if err != nil {
			return meta, err
		}
		content, err := markdown2html.Convert(markdownContent) // Convert markdown to html and assign it to `.content`
		if err != nil {
			return meta, err
		}
		meta["content"] = string(content)
	}

	// with .<values>
	for key, value := range engine.Values {
		meta[key] = value
	}

	return meta, nil
}

// createBreadcrumbs creates breadcrumb structs from a rendered template path
// Breadcrumbs represent parent directories, excluding the directory containing the index.html
// Examples:
//   - "index.html" -> [] (empty)
//   - "a/index.html" -> [] (empty, no parent)
//   - "a/b/index.html" -> [{Name: "a", Path: "/a"}]
//   - "a/b/c/index.html" -> [{Name: "a", Path: "/a"}, {Name: "b", Path: "/a/b"}]
func createBreadcrumbs(renderedTemplatePath string) []Breadcrumb {
	// Remove filename: "a/b/c/index.html" -> "a/b/c"
	templateDir := path.Dir(renderedTemplatePath)

	// Go one folder up to get parent directory: "a/b/c" -> "a/b"
	parentDir := path.Dir(templateDir)

	// If parentDir is "." (root), return empty breadcrumbs
	// This handles cases like "index.html" and "a/index.html"
	if parentDir == "." {
		return []Breadcrumb{}
	}

	// Split into directory names
	dirNames := strings.Split(parentDir, "/")

	// Filter out empty strings and "." (root)
	var cleanDirNames []string
	for _, name := range dirNames {
		if name != "" && name != "." {
			cleanDirNames = append(cleanDirNames, name)
		}
	}

	// If no breadcrumbs, return empty slice
	if len(cleanDirNames) == 0 {
		return []Breadcrumb{}
	}

	// Build breadcrumbs with accumulated paths
	breadcrumbs := []Breadcrumb{}
	currentPath := ""

	for _, dirName := range cleanDirNames {
		// Build path incrementally
		if currentPath == "" {
			currentPath = "/" + dirName
		} else {
			currentPath = currentPath + "/" + dirName
		}

		breadcrumb := Breadcrumb{
			Name: dirName,
			Path: currentPath,
		}
		breadcrumbs = append(breadcrumbs, breadcrumb)
	}

	return breadcrumbs
}
