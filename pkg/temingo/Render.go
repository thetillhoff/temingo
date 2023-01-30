package temingo

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/thetillhoff/temingo/pkg/fileIO"
)

func (engine *Engine) Render() error {
	var (
		err         error
		ignoreLines []string
		fileList    fileIO.FileList

		componentPaths    []string
		templatePaths     []string
		metaTemplatePaths []string
		staticPaths       []string

		content              []byte
		renderedTemplatePath string

		componentFiles        = map[string]string{}
		renderedTemplates     = map[string][]byte{}
		renderedMetaTemplates map[string][]byte
	)

	// Parse temingoignore if exists

	if _, err = os.Stat(engine.TemingoignorePath); os.IsNotExist(err) {
		// No ignore file
	} else if err != nil {
		// ignore file exists, but can't be accessed
		return err
	} else {
		// temingoignore exists and can be read

		ignoreLines, err = fileIO.ReadFileLineByLine(engine.TemingoignorePath)
		if err != nil {
			return err
		}
	}

	// Read filetree with ignoreLines

	fileList = fileIO.FileList{Path: engine.InputDir}

	err = fileList.GenerateWithIgnoreLines(ignoreLines, engine.Verbose) // Get inputDir file-tree
	if err != nil {
		return err
	}

	// Sort retrieved filepaths

	templatePaths, metaTemplatePaths, componentPaths, staticPaths = engine.sortPaths(fileList)

	// Read component files

	for _, componentPath := range componentPaths { // Read contents of each component file
		content, err = fileIO.ReadFile(path.Join(engine.InputDir, componentPath)) // Read file contents
		if err != nil {
			return err
		}

		componentFiles[componentPath] = string(content)
	}

	// Verify components

	err = verifyComponents(componentFiles) // Check if the components are unique
	if err != nil {
		return err
	}

	// Read template files and execute them

	for _, templatePath := range templatePaths {
		content, err = fileIO.ReadFile(path.Join(engine.InputDir, templatePath))
		if err != nil {
			return err
		}

		// TODO move getMetaForDir here (currently in renderTemplate())
		// it should return two maps; meta and childMeta
		// OR it should return one map, where map["meta"] and map["childMeta"] are already set

		renderedTemplatePath = strings.ReplaceAll(templatePath, engine.TemplateExtension, "")
		renderedTemplates[renderedTemplatePath], err = engine.renderTemplate(fileList, renderedTemplatePath, string(content), componentFiles) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
		if err != nil {
			return err
		}
	}

	// Read metatemplate files, check metadata and execute them

	for _, metaTemplatePath := range metaTemplatePaths { // Read metaTemplate contents and execute them for each childfolder that contains a meta.yaml
		content, err = fileIO.ReadFile(path.Join(engine.InputDir, metaTemplatePath))
		if err != nil {
			return err
		}

		renderedMetaTemplates, err = engine.renderMetaTemplate(fileList, metaTemplatePath, string(content), componentFiles) // There will be multiple rendered files out of one meta template
		if err != nil {
			return err
		}
		for renderedTemplatePath, content = range renderedMetaTemplates {
			renderedTemplates[renderedTemplatePath] = content
		}
	}

	// Update output

	if !engine.DryRun { // Only if dry-run is disabled

		err = os.RemoveAll(engine.OutputDir) // Ensure the outputDir is empty
		if err != nil {
			return err
		}
		err = fileIO.CopyFile(engine.InputDir, engine.OutputDir) // Recreate the outputDir with the same permissions as the inputDir
		if err != nil {
			return err
		}

		for _, staticPath := range staticPaths {
			err = fileIO.CopyFile(path.Join(engine.InputDir, staticPath), path.Join(engine.OutputDir, staticPath))
			if err != nil {
				return err
			}
			if engine.Verbose {
				log.Println("Writing static file to " + path.Join(engine.OutputDir, staticPath))
			}
		}

		for templatePath, renderedTemplate := range renderedTemplates { // includes both templates and metaTemplates
			err = fileIO.WriteFile(path.Join(engine.OutputDir, templatePath), renderedTemplate)
			if err != nil {
				return err
			}
			if engine.Verbose {
				log.Println("Writing rendered template to " + path.Join(engine.OutputDir, templatePath))
			}
		}
	}

	return nil
}
