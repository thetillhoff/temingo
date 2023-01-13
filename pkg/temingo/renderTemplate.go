package temingo

import (
	"bytes"
	"text/template"
)

// Returns the rendered template
func renderTemplate(templatePath string, templateContent string, componentFiles map[string]string) ([]byte, error) {
	var (
		err          error
		values       map[string]interface{} = make(map[string]interface{})
		outputBuffer *bytes.Buffer          = new(bytes.Buffer)
	)

	// Create Values object
	// with breadcrumbs, template name, ...
	values["path"] = templatePath

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
