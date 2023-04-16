package temingo

import (
	"bytes"
	"text/template"
)

// Returns the rendered template
func (engine *Engine) renderTemplate(meta map[string]interface{}, templatePath string, templateContent string, partialFiles map[string]string) ([]byte, error) {
	var (
		err error

		outputBuffer *bytes.Buffer = new(bytes.Buffer)
	)

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
