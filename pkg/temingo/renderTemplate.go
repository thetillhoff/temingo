package temingo

import (
	"bytes"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

// Returns the rendered template
func renderTemplate(templatePath string, templateContent string, componentFiles map[string]string, inputDir string) ([]byte, error) {
	var (
		err             error
		values          map[string]interface{} = make(map[string]interface{})
		templateDir     string
		childMetaForDir map[string]interface{}
		childMeta       map[string]interface{} = make(map[string]interface{})
		outputBuffer    *bytes.Buffer          = new(bytes.Buffer)
	)

	// Create Values object
	// with breadcrumbs, template name, ...
	values["path"] = templatePath // Path to the current file (without `src/` or `output/`)

	templateDir, _ = path.Split(templatePath)
	values["breadcrumbs"] = strings.Split(templateDir, "/") // Breadcrumbs to the current file

	log.Println("Searching metadata for", templatePath)
	values["meta"], err = getMetaForDir(inputDir, templateDir) // Contains aggregated `meta.yaml`s (up to parent dir, were children overwrite their parents values during the merge)
	if err != nil {
		return nil, err
	}

	templateDir, _ = path.Split(templatePath)
	files, err := os.ReadDir(path.Join(inputDir, templateDir)) // Get all child-elements of folder
	if err != nil {
		return nil, err
	}
	for _, f := range files { // For each child-element of folder
		if f.IsDir() { // Only for folders
			log.Println("Searching child metadata for", inputDir+path.Join(path.Dir(templatePath), f.Name()))
			childMetaForDir, err = getMetaForDir(inputDir, path.Join(path.Dir(templatePath), f.Name()))
			if err != nil {
				return nil, err
			}
			if len(childMetaForDir) > 0 {
				childMeta[f.Name()] = childMetaForDir
			}
		}
	}
	values["childmeta"] = childMeta

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
	err = templateEngine.Execute(outputBuffer, values)
	if err != nil {
		return nil, err
	}

	// return template
	return outputBuffer.Bytes(), nil
}
