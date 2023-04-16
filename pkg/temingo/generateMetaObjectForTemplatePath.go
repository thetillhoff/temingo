package temingo

import (
	"log"
	"path"
	"strings"

	"github.com/thetillhoff/fileIO"
	"github.com/thetillhoff/temingo/pkg/markdown2html"
)

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
	templateDir, _ := path.Split(renderedTemplatePath)
	meta["breadcrumbs"] = strings.Split(templateDir, "/") // Breadcrumbs to the current file

	// with .meta and .childMeta
	if engine.Verbose {
		log.Println("Searching metadata for", renderedTemplatePath)
	}
	meta["meta"], meta["childMeta"], err = engine.getMetaForTemplatePath(fileIO.FileList{Files: metaPaths}, renderedTemplatePath) // Contains aggregated meta yamls (up to parent dir, were children overwrite their parents values during the merge)
	if err != nil {
		return meta, err
	}

	// with .content
	markdownContentFiles := fileList.FilterByFolderPath(path.Dir(templatePath)).FilterByFilename(engine.MarkdownContentFilename).Files
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

	return meta, nil

}
