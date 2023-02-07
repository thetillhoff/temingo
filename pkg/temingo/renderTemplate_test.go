package temingo

// // Check if valid template without partial references renders successfully and to the expected result
// func TestRenderTemplateWithValidTemplate(t *testing.T) {
// 	fileList := fileIO.FileList{
// 		Files: []string{},
// 	}
// 	templatePath := "random/path/test.template.txt"
// 	templateContent := `{{ .path }}`
// 	partialFiles := map[string]string{}

// 	expectedValue := `random/path/test.template.txt`

// 	engine := DefaultEngine()

// 	renderedTemplate, err := engine.renderTemplate(fileList, templatePath, templateContent, partialFiles)
// 	if err != nil {
// 		t.Fatal("expected template rendering to be successful, got error:", err)
// 	}

// 	if string(renderedTemplate) != expectedValue {
// 		t.Fatal("expected template content to be", expectedValue, "but got", string(renderedTemplate))
// 	}

// }

// // Check if valid template with partial references renders successfully and to the expected result
// func TestRenderTemplateWithValidTemplateAndPartialFile(t *testing.T) {
// 	fileList := fileIO.FileList{
// 		Files: []string{},
// 	}
// 	templatePath := "random/path/test.template.txt"
// 	templateContent := `{{ template "partialA" }}`
// 	expectedValue := `test`
// 	partialFiles := map[string]string{
// 		"path does not matter here": `{{ define "partialA" }}
// test
// {{ end }}
// `,
// 	}

// 	engine := DefaultEngine()

// 	renderedTemplate, err := engine.renderTemplate(fileList, templatePath, templateContent, partialFiles)
// 	if err != nil {
// 		t.Fatal("expected template rendering to be successful, got error:", err)
// 	}

// 	if string(renderedTemplate) != expectedValue {
// 		t.Fatal("exected rendered template to be", expectedValue, "but got", string(renderedTemplate))
// 	}

// }

// TODO add test with meta data

// TODO add test with parent meta data (multiple meta data points)

// TODO add test with invalid template
