package temingo

import (
	"bytes"
	"text/template"
)

// Returns the rendered template
func (engine *Engine) renderTemplate(meta map[string]interface{}, templatePath string, templateContent string, partialFiles map[string]string) ([]byte, error) {
	logger := engine.Logger

	var (
		err error

		outputBuffer *bytes.Buffer = new(bytes.Buffer)
	)

	logger.Debug("Meta object for template", "path", templatePath, "meta", meta)

	outputBuffer.Reset()                         // Ensure the buffer is empty
	templateEngine := template.New(templatePath) // Create a new template with the path to it as its name

	// Defining additional template functions
	templateEngine = templateEngine.Funcs(template.FuncMap{
		"concat":                 tmpl_concat,
		"includeWithIndentation": tmpl_indent,
		"capitalize":             tmpl_capitalize,
	})

	for _, partialFileContent := range partialFiles { // For each partialFile
		templateEngine.Parse(partialFileContent) // Parse the partials contained in it
	}

	_, err = templateEngine.Parse(templateContent) // Parse the template
	if err != nil {
		return nil, err
	}

	err = templateEngine.Execute(outputBuffer, meta)
	if err != nil {
		return nil, err
	}

	// return template
	return outputBuffer.Bytes(), nil
}
