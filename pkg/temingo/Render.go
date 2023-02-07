package temingo

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/thetillhoff/temingo/pkg/fileIO"
	prettifyhtml "github.com/thetillhoff/temingo/pkg/prettifyHTML"
)

// Renders the templates in the inputDir, writes them to the outputDir and copies the static files
func (engine *Engine) Render() error {
	var (
		err         error
		ignoreLines []string
		fileList    fileIO.FileList

		partialPaths      []string
		templatePaths     []string
		metaTemplatePaths []string
		metaPaths         []string
		staticPaths       []string

		content              []byte
		renderedTemplatePath string

		partialFiles      = map[string]string{}
		renderedTemplates = map[string][]byte{}
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

	fileList, err = fileIO.GenerateFileListWithIgnoreLines(engine.InputDir, ignoreLines, engine.Verbose) // Get inputDir file-tree
	if err != nil {
		return err
	}
	if len(fileList.Files) == 0 {
		log.Println("There were no files found in", fileList.Path)
	}

	// Sort retrieved filepaths

	templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths = engine.sortPaths(fileList)

	// Read partial files

	for _, partialPath := range partialPaths { // Read contents of each partial file
		content, err = fileIO.ReadFile(path.Join(engine.InputDir, partialPath)) // Read file contents
		if err != nil {
			return err
		}

		partialFiles[partialPath] = string(content)
	}

	// Verify partials

	err = verifyPartials(partialFiles) // Check if the partials are unique
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

		// TODO renderTemplate should get the full meta/values object passed, without having to access the filesystem

		renderedTemplatePath = strings.ReplaceAll(templatePath, engine.TemplateExtension, "")
		renderedTemplates[renderedTemplatePath], err = engine.renderTemplate(fileIO.FileList{Files: metaPaths}, templatePath, string(content), partialFiles) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
		if err != nil {
			return err
		}

		switch path.Ext(renderedTemplatePath) { // File type autodetection
		case ".html":
			// Prettify by default
			renderedTemplates[renderedTemplatePath] = []byte(prettifyhtml.Format(string(renderedTemplates[renderedTemplatePath]))) // Meh about the conversions
		}
	}

	// Read metatemplate files, check metadata and execute them

	for _, metaTemplatePath := range metaTemplatePaths { // Read metaTemplate contents and execute them for each childfolder that contains a meta yaml
		content, err = fileIO.ReadFile(path.Join(engine.InputDir, metaTemplatePath))
		if err != nil {
			return err
		}

		for _, metaFilePath := range fileList.FilterByLevelAtFolderPath(path.Dir(metaTemplatePath), 1).FilterByFilename(engine.MetaFilename).Files { // For each meta yaml in a direct subfolder
			if engine.Verbose {
				log.Println("Found metatemplate child at", metaFilePath)
			}

			renderedTemplatePath = path.Join(path.Dir(metaFilePath), path.Base(metaTemplatePath))             // == Location of meta yaml, minus meta yaml, plus filename of metatemplate
			renderedTemplatePath = strings.ReplaceAll(renderedTemplatePath, engine.MetaTemplateExtension, "") // Remove template extension from filename

			renderedTemplates[renderedTemplatePath], err = engine.renderTemplate(fileIO.FileList{Files: metaPaths}, renderedTemplatePath, string(content), partialFiles) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
			if err != nil {
				return err
			}

			switch path.Ext(renderedTemplatePath) { // File type autodetection
			case ".html":
				// Prettify by default
				renderedTemplates[renderedTemplatePath] = []byte(prettifyhtml.Format(string(renderedTemplates[renderedTemplatePath]))) // Meh about the conversions
			}

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
	} else { // DryRun, so provide information about what would be done instead of doing it
		log.Println("Would write the following", len(renderedTemplates), "rendered templates:")
		for templatePath := range renderedTemplates {
			log.Println("-", templatePath)
		}

		log.Println("Would write the following", len(staticPaths), "static files:")
		for _, staticPath := range staticPaths {
			log.Println("-", staticPath)
		}
	}

	return nil
}
