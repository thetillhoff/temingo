package temingo

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

func Render(inputDir string, outputDir string, temingoignorePath string) error {
	var (
		err       error
		filePaths []string

		componentPaths    []string
		templatePaths     []string
		metatemplatePaths []string
		staticPaths       []string

		temporaryTemplateEngineName = "temporaryComponentEngine"
		content                     []byte
		temporaryTemplateEngine     *template.Template
		componentName               string
		componentLocations          = make(map[string]string)
		renderedTemplate            []byte
		metaTemplateRenderPath      string

		componentExtension    = ".component"
		templateExtension     = ".template"
		metaTemplateExtension = ".metatemplate"
		componentFiles        = make(map[string]string)
		renderedTemplates     = make(map[string][]byte)
	)

	filePaths, err = retrieveFilePaths(inputDir, temingoignorePath) // Get inputDir file-tree
	if err != nil {
		return err
	}

	for _, filePath := range filePaths { // Check what type of file we have
		if strings.Contains(filePath, componentExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			componentPaths = append(componentPaths, filePath)
			log.Println("Identified as component file:", filePath)
		} else if strings.Contains(filePath, templateExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			templatePaths = append(templatePaths, filePath)
			log.Println("Identified as template file:", filePath)
		} else if strings.Contains(filePath, metaTemplateExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			metatemplatePaths = append(metatemplatePaths, filePath)
			log.Println("Identified as metatemplate file:", filePath)
		} else {
			staticPaths = append(staticPaths, filePath)
			log.Println("Identified as static file:", filePath)
		}
	}

	for _, componentPath := range componentPaths { // For each component filepath
		content, err = readFile(path.Join(inputDir, componentPath)) // Read file contents
		if err != nil {
			return err
		}

		// Checking for duplicate components
		temporaryTemplateEngine = template.New(temporaryTemplateEngineName) // Create a new temporary template
		_, err = temporaryTemplateEngine.Parse(string(content))             // Parse the component into the temporary template engine
		if err != nil {
			return err
		}
		componentName = strings.TrimPrefix(temporaryTemplateEngine.DefinedTemplates(), "; defined templates are: ") // Prefix comes from the offical text.template library
		componentName = strings.ReplaceAll(componentName, "\"", "")                                                 // remove '"'
		componentName = strings.ReplaceAll(componentName, " ", "")                                                  // remove ' '
		for _, subcomponentName := range strings.Split(componentName, ",") {                                        // For all components in this component file (check if it's name is unique)
			if subcomponentName == temporaryTemplateEngineName { // Skip the manually added initial template engine name
				continue
			} else {
				for existingComponentName, existingComponentPath := range componentLocations { // For each component that already exists
					if subcomponentName == existingComponentName { // If new component would overwrite an existing component (==same name)
						return errors.New("duplicate component name '" + subcomponentName + "' found in " + componentPath + " and " + existingComponentPath)
					}
				}
				// If the component is truly new
				componentLocations[subcomponentName] = componentPath // Add the component name to the list. ComponentPath is only used to provide a better error message
			}
		}
		// All components in this component file are unique
		componentFiles[componentPath] = string(content) // Add the component file contents to the component file list
	}

	for _, templatePath := range templatePaths { // Read template contents and execute them
		content, err = readFile(path.Join(inputDir, templatePath))
		if err != nil {
			return err
		}

		templateRenderPath = strings.ReplaceAll(templatePath, templateExtension, "")
		renderedTemplate, err = renderTemplate(templateRenderPath, string(content), componentFiles, inputDir) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
		if err != nil {
			return err
		}

		renderedTemplates[templateRenderPath] = renderedTemplate
	}

	for _, metaTemplatePath := range metatemplatePaths { // Read metaTemplate contents and execute them for each childfolder that contains a meta.yaml
		content, err = readFile(path.Join(inputDir, metaTemplatePath))
		if err != nil {
			return err
		}

		files, err := ioutil.ReadDir(path.Dir(metaTemplatePath)) // Get all child-elements of folder
		if err != nil {
			return err
		}

		for _, f := range files { // For each child-element of folder
			if f.IsDir() { // Only for folders
				if _, err = os.Stat(path.Join(metaTemplatePath, f.Name(), "meta.yaml")); !os.IsNotExist(err) { // Check if meta.yaml exists
					metaTemplateRenderPath = strings.ReplaceAll(path.Join(path.Dir(metaTemplatePath), f.Name(), path.Base(metaTemplatePath)), metaTemplateExtension, "")
					renderedTemplate, err = renderTemplate(metaTemplateRenderPath, string(content), componentFiles, inputDir) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
					if err != nil {
						return err
					}
				}
			}
		}

		renderedTemplates[metaTemplateRenderPath] = renderedTemplate
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
		log.Println("writing static file to " + path.Join(outputDir, staticPath))
	}

	for templatePath, renderedTemplate := range renderedTemplates {
		err = writeFile(path.Join(outputDir, templatePath), renderedTemplate)
		if err != nil {
			return err
		}
		log.Println("writing rendered template to " + path.Join(outputDir, templatePath))
	}

	log.Println("Build completed.")

	return nil
}
