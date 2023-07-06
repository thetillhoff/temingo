package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
)

func runTemplate(mappedValues map[string]interface{}, templateName string, template string, partialTemplates [][]string, outputFilePath string) {
	outputBuffer := new(bytes.Buffer)
	outputBuffer.Reset()
	tpl := parseTemplateFiles(templateName, template, partialTemplates)
	mappedValues["breadcrumbs"] = createBreadcrumbs(filepath.Dir(templateName))
	err := tpl.Execute(outputBuffer, mappedValues)
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := os.Stat(outputDir); os.IsNotExist(err) { // If output directory doesn't exist
		createFolderIfNotExists(outputDir)
	}
	err = writeTemplateToFile(outputFilePath, outputBuffer.Bytes())
	if err != nil {
		log.Fatalln(err)
	}
}
