package temingo

import (
	"bytes"
	"log"
	"path"
	"strings"
	"text/template"

	"github.com/thetillhoff/temingo/pkg/fileIO"
)

// Returns the rendered template
func (engine *Engine) renderTemplate(metaTemplatePaths fileIO.FileList, templatePath string, templateContent string, componentFiles map[string]string) ([]byte, error) {
	var (
		err         error
		meta        map[string]interface{} = map[string]interface{}{}
		templateDir string

		outputBuffer *bytes.Buffer = new(bytes.Buffer)
	)

	// Create Values object
	// with breadcrumbs, template name, ...
	meta["path"] = templatePath // Path to the current file (without `src/` or `output/`)

	templateDir, _ = path.Split(templatePath)
	meta["breadcrumbs"] = strings.Split(templateDir, "/") // Breadcrumbs to the current file

	if engine.Verbose {
		log.Println("Searching metadata for", templatePath)
	}

	meta, err = engine.getMetaForTemplatePath(metaTemplatePaths, templatePath) // Contains aggregated `meta.yaml`s (up to parent dir, were children overwrite their parents values during the merge)
	if err != nil {
		return nil, err
	}

	outputBuffer.Reset()                         // Ensure the buffer is empty
	templateEngine := template.New(templatePath) // Create a new template with the path to it as its name

	for _, componentFileContent := range componentFiles { // For each componentFile
		templateEngine.Parse(componentFileContent) // Parse the components contained in it
	}

	// tpl.Funcs(funcMap).Parse(baseTemplate)
	_, err = templateEngine.Parse(templateContent) // Parse the template
	if err != nil {
		return nil, err
	}
	// TODO Template functionmap is missing here
	err = templateEngine.Execute(outputBuffer, meta)
	if err != nil {
		return nil, err
	}

	// return template
	return outputBuffer.Bytes(), nil
}
