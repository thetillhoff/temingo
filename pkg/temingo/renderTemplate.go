package temingo

import (
	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"text/template"
)

// Returns the rendered template
func renderTemplate(templatePath string, templateContent string, componentFiles map[string]string, inputDir string) ([]byte, error) {
	var (
		err          error
		values       map[string]interface{} = make(map[string]interface{})
		childMeta    map[string]interface{} = make(map[string]interface{})
		outputBuffer *bytes.Buffer          = new(bytes.Buffer)
	)

	// Create Values object
	// with breadcrumbs, template name, ...
	values["path"] = templatePath // Path to the current file (without `src/` or `output/`)

	values["breadcrumbs"] = strings.Split(path.Dir(templatePath), "/") // Breadcrumbs to the current file

	values["meta"], err = getMetaForDir(path.Dir(templatePath), inputDir) // Contains aggregated `meta.yaml`s (up to parent dir, were children overwrite their parents values during the merge)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(path.Dir(templatePath)) // Get all child-elements of folder
	if err != nil {
		return nil, err
	}
	for _, f := range files { // For each child-element of folder
		if f.IsDir() { // Only for folders
			childMeta[f.Name()], err = getMetaForDir(path.Join(path.Dir(templatePath), f.Name()), inputDir)
			if err != nil {
				return nil, err
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
	templateEngine.Parse(templateContent) // Parse the template
	// TODO Template functionmap is missing here
	err = templateEngine.Execute(outputBuffer, values)

	// return template
	return outputBuffer.Bytes(), err
}
