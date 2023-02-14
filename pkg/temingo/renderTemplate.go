package temingo

import (
	"bytes"
	"log"
	"path"
	"strings"
	"text/template"

	"github.com/thetillhoff/fileIO/v2"
)

// Returns the rendered template
func (engine *Engine) renderTemplate(metaTemplatePaths fileIO.FileList, templatePath string, templateContent string, partialFiles map[string]string) ([]byte, error) {
	var (
		err       error
		meta      map[string]interface{} = map[string]interface{}{}
		fileMeta  interface{}
		childMeta map[string]interface{}

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

	fileMeta, childMeta, err = engine.getMetaForTemplatePath(metaTemplatePaths, templatePath) // Contains aggregated meta yamls (up to parent dir, were children overwrite their parents values during the merge)
	if err != nil {
		return nil, err
	}

	meta["meta"] = fileMeta
	meta["childMeta"] = childMeta

	outputBuffer.Reset()                         // Ensure the buffer is empty
	templateEngine := template.New(templatePath) // Create a new template with the path to it as its name

	for _, partialFileContent := range partialFiles { // For each partialFile
		templateEngine.Parse(partialFileContent) // Parse the partials contained in it
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
