package temingo

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/thetillhoff/fileIO"
)

// Renders the templates in the inputDir, writes them to the outputDir and copies the static files
func (engine *Engine) Render() error {
	logger := engine.Logger

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
		templateContents  []string
	)

	if err = engine.validateEngine(); err != nil {
		return err
	}

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
			return fmt.Errorf("reading temingoignore: %w", err)
		}
	}

	// Add values files to ignore lines if specified
	for _, valuesFilePath := range engine.ValuesFilePaths {
		ignoreLines = append(ignoreLines, valuesFilePath)
		logger.Debug("Adding values file to ignore list", "path", valuesFilePath)
	}

	// Validate directories and get ignore path in one call
	// This validates directories exist, creates outputDir if needed, and calculates the ignore path
	outputIgnorePath, err := validateDirectories(engine.InputDir, engine.OutputDir, engine.NoDeleteOutputDir, logger)
	if err != nil {
		return err
	}
	if outputIgnorePath != "" {
		logger.Warn("Output directory is inside input directory. Adding to ignore list to prevent processing loops", "path", outputIgnorePath)
		ignoreLines = append(ignoreLines, outputIgnorePath)
		logger.Debug("Adding output directory to ignore list", "path", outputIgnorePath)
	}

	// Read filetree with ignoreLines
	fileList, err = fileIO.GenerateFileListWithIgnoreLines(engine.InputDir, ignoreLines, engine.Verbose)
	if err != nil {
		return fmt.Errorf("reading input directory %s: %w", engine.InputDir, err)
	}
	if len(fileList.Files) == 0 {
		logger.Warn("No files found in input directory", "path", fileList.Path)
	}

	if err = validateFolderNames(fileList); err != nil {
		return err
	}

	// Sort retrieved filepaths
	templatePaths, metaTemplatePaths, partialPaths, metaPaths, _, staticPaths = engine.sortPaths(fileList) // markdown content files are picked up later anyway

	// Read partial files
	for _, partialPath := range partialPaths {
		content, err = fileIO.ReadFile(path.Join(engine.InputDir, partialPath))
		if err != nil {
			return fmt.Errorf("reading partial %s: %w", partialPath, err)
		}
		partialFiles[partialPath] = "{{ define \"" + partialPath + "\" -}}\n" + string(content) + "\n{{- end -}}"
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
			return fmt.Errorf("reading template %s: %w", templatePath, err)
		}

		templateContents = append(templateContents, string(content))
		renderedTemplatePath = strings.ReplaceAll(templatePath, engine.TemplateExtension, "")

		// Create meta values object
		meta, err := engine.generateMetaObjectForTemplatePath(renderedTemplatePath, fileList, metaPaths)
		if err != nil {
			return err
		}

		renderedTemplates[renderedTemplatePath], err = engine.renderTemplate(meta, templatePath, string(content), partialFiles) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
		if err != nil {
			return fmt.Errorf("rendering template %s: %w", templatePath, err)
		}
	}

	// Read metatemplate files, check metadata & markdown content files and execute them
	for _, metaTemplatePath := range metaTemplatePaths { // Read metaTemplate contents and execute them for each childfolder that contains a meta yaml
		content, err = fileIO.ReadFile(path.Join(engine.InputDir, metaTemplatePath))
		if err != nil {
			return fmt.Errorf("reading metatemplate %s: %w", metaTemplatePath, err)
		}
		templateContents = append(templateContents, string(content))

		for _, metaFilePath := range fileList.FilterByLevelAtFolderPath(path.Dir(metaTemplatePath), 1).FilterByFilename(engine.MetaFilename).Files { // For each meta yaml in a direct subfolder
			logger.Debug("Found metatemplate child", "path", metaFilePath)

			renderedTemplatePath = path.Join(path.Dir(metaFilePath), path.Base(metaTemplatePath))             // == Location of meta yaml, minus meta yaml, plus filename of metatemplate
			renderedTemplatePath = strings.ReplaceAll(renderedTemplatePath, engine.MetaTemplateExtension, "") // Remove template extension from filename

			// Create meta values object
			meta, err := engine.generateMetaObjectForTemplatePath(renderedTemplatePath, fileList, metaPaths)
			if err != nil {
				return err
			}

			renderedTemplates[renderedTemplatePath], err = engine.renderTemplate(meta, renderedTemplatePath, string(content), partialFiles) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
			if err != nil {
				return fmt.Errorf("rendering metatemplate %s for %s: %w", metaTemplatePath, renderedTemplatePath, err)
			}
		}
	}

	engine.warnUnusedPartials(partialFiles, templateContents)

	// Beautify/Minify

	if engine.Beautify {
		for renderedTemplatePath, content := range renderedTemplates {
			renderedTemplates[renderedTemplatePath] = engine.beautify(content, path.Ext(renderedTemplatePath))
		}
	} else if engine.Minify {
		for renderedTemplatePath, content := range renderedTemplates {
			renderedTemplates[renderedTemplatePath] = engine.minify(content, path.Ext(renderedTemplatePath))
		}
	}

	// Warn on insecure http:// links in rendered output
	for renderedTemplatePath, content := range renderedTemplates {
		if path.Ext(renderedTemplatePath) == ".html" {
			engine.warnHTTPLinks(renderedTemplatePath, content)
		}
	}

	// Update output
	if !engine.DryRun { // Only if dry-run is disabled
		if !engine.NoDeleteOutputDir {
			err = os.RemoveAll(engine.OutputDir) // Ensure the outputDir is empty
			if err != nil {
				return fmt.Errorf("clearing output directory %s: %w", engine.OutputDir, err)
			}
			err = fileIO.CopyFile(engine.InputDir, engine.OutputDir) // Recreate the outputDir with the same permissions as the inputDir
			if err != nil {
				return fmt.Errorf("recreating output directory %s: %w", engine.OutputDir, err)
			}

			// Ensure output directory permissions match input directory (fileIO.CopyFile may not preserve them)
			inputDirInfo, err := os.Stat(engine.InputDir)
			if err != nil {
				return fmt.Errorf("error getting input directory info: %w", err)
			}
			outputDirToCheck := strings.TrimSuffix(engine.OutputDir, string(filepath.Separator))
			if outputDirToCheck == "" {
				outputDirToCheck = engine.OutputDir
			}
			if err := os.Chmod(outputDirToCheck, inputDirInfo.Mode().Perm()); err != nil {
				return fmt.Errorf("error setting output directory permissions: %w", err)
			}
			logger.Debug("Recreating output directory", "path", engine.OutputDir)

			for _, staticPath := range staticPaths {
				err = fileIO.CopyFile(path.Join(engine.InputDir, staticPath), path.Join(engine.OutputDir, staticPath))
				if err != nil {
					return fmt.Errorf("copying static file %s: %w", staticPath, err)
				}
				logger.Debug("Writing static file", "path", path.Join(engine.OutputDir, staticPath))
			}
		}

		for templatePath, renderedTemplate := range renderedTemplates { // includes both templates and metaTemplates

			if engine.NoDeleteOutputDir {
				if _, err = os.Stat(path.Join(engine.OutputDir, templatePath)); err == nil {
					err = os.Remove(path.Join(engine.OutputDir, templatePath))
					if err != nil {
						return fmt.Errorf("removing existing output file %s: %w", path.Join(engine.OutputDir, templatePath), err)
					}
					logger.Debug("Deleting existing rendered template", "path", path.Join(engine.OutputDir, templatePath))
				}
			}

			// Get permissions from input directory (used for both files and parent directories)
			// For template files, we use input directory permissions as the source of truth
			inputDirInfo, err := os.Stat(engine.InputDir)
			if err != nil {
				return fmt.Errorf("error getting input directory info: %w", err)
			}
			fileMode := inputDirInfo.Mode().Perm()

			// Ensure parent directory exists with same permissions as input directory
			outputFilePath := path.Join(engine.OutputDir, templatePath)
			outputDirPath := path.Dir(outputFilePath)
			if err := os.MkdirAll(outputDirPath, fileMode); err != nil {
				return fmt.Errorf("error creating output directory %s: %w", outputDirPath, err)
			}

			// Use Chmod to ensure exact permissions (MkdirAll may be affected by umask)
			if err := os.Chmod(outputDirPath, fileMode); err != nil {
				return fmt.Errorf("error setting output directory permissions %s: %w", outputDirPath, err)
			}

			err = fileIO.WriteFile(outputFilePath, renderedTemplate)
			if err != nil {
				return fmt.Errorf("writing output file %s: %w", outputFilePath, err)
			}

			// Set file permissions to match input directory permissions
			if err := os.Chmod(outputFilePath, fileMode); err != nil {
				return fmt.Errorf("error setting permissions for %s: %w", outputFilePath, err)
			}

			logger.Debug("Writing rendered template", "path", outputFilePath)
		}
	} else { // DryRun, so provide information about what would be done instead of doing it
		logger.Info("Dry run: would write rendered templates", "count", len(renderedTemplates))
		for templatePath := range renderedTemplates {
			logger.Debug("Would write rendered template", "path", templatePath)
		}

		logger.Info("Dry run: would write static files", "count", len(staticPaths))
		for _, staticPath := range staticPaths {
			logger.Debug("Would write static file", "path", staticPath)
		}
	}

	return nil
}
