package temingo

import (
	"log"
	"os"
	"path"
	"strings"
)

func Render(inputDirFlag string, outputDirFlag string, temingoignorePathFlag string, templateExtensionFlag string, metaTemplateExtensionFlag string, componentExtensionFlag string, verboseFlag bool) error {
	var (
		err       error
		filePaths []string

		componentPaths    []string
		templatePaths     []string
		metaTemplatePaths []string
		staticPaths       []string

		content              []byte
		renderedTemplatePath string

		componentFiles        = make(map[string]string)
		renderedTemplates     = make(map[string][]byte)
		renderedMetaTemplates map[string][]byte
	)

	// Set flags globally so they don't have to be passed around all the time
	inputDir = inputDirFlag
	outputDir = outputDirFlag
	temingoignorePath = temingoignorePathFlag
	templateExtension = templateExtensionFlag
	metaTemplateExtension = metaTemplateExtensionFlag
	componentExtension = componentExtensionFlag
	verbose = verboseFlag

	filePaths, err = retrieveFilePaths() // Get inputDir file-tree
	if err != nil {
		return err
	}

	templatePaths, metaTemplatePaths, componentPaths, staticPaths = sortPaths(filePaths)

	for _, componentPath := range componentPaths { // Read contents of each component file
		content, err = readFile(path.Join(inputDir, componentPath)) // Read file contents
		if err != nil {
			return err
		}

		componentFiles[componentPath] = string(content)
	}

	err = verifyComponents(componentFiles) // Check if the components are unique
	if err != nil {
		return err
	}

	for _, templatePath := range templatePaths { // Read template contents and execute them
		content, err = readFile(path.Join(inputDir, templatePath))
		if err != nil {
			return err
		}

		renderedTemplatePath = strings.ReplaceAll(templatePath, templateExtension, "")
		renderedTemplates[renderedTemplatePath], err = renderTemplate(renderedTemplatePath, string(content), componentFiles) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
		if err != nil {
			return err
		}
	}

	for _, metaTemplatePath := range metaTemplatePaths { // Read metaTemplate contents and execute them for each childfolder that contains a meta.yaml
		content, err = readFile(path.Join(inputDir, metaTemplatePath))
		if err != nil {
			return err
		}

		renderedMetaTemplates, err = renderMetaTemplate(metaTemplatePath, string(content), componentFiles) // There will be multiple rendered files out of one meta template
		for renderedTemplatePath, content = range renderedMetaTemplates {
			renderedTemplates[renderedTemplatePath] = content
		}
	}

	err = os.RemoveAll(outputDir) // Ensure the outputDir is empty
	if err != nil {
		return err
	}
	err = copyFile(inputDir, outputDir) // Recreate the outputDir with the same permissions as the inputDir
	if err != nil {
		return err
	}

	for _, staticPath := range staticPaths {
		err = copyFile(path.Join(inputDir, staticPath), path.Join(outputDir, staticPath))
		if err != nil {
			return err
		}
		if verbose {
			log.Println("Writing static file to " + path.Join(outputDir, staticPath))
		}
	}

	for templatePath, renderedTemplate := range renderedTemplates {
		err = writeFile(path.Join(outputDir, templatePath), renderedTemplate)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("Writing rendered template to " + path.Join(outputDir, templatePath))
		}
	}

	return nil
}
