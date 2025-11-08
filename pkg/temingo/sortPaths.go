package temingo

import (
	"path"
	"slices"
	"strings"

	"github.com/thetillhoff/fileIO"
)

// Takes the paths from FileList.Files and sorts them into one list per filetype
// Order of returned lists: templatePaths, metaTemplatePaths, partialPaths, metaPaths, staticPaths
func (engine *Engine) sortPaths(fileList fileIO.FileList) ([]string, []string, []string, []string, []string, []string) {
	var (
		templatePaths        []string
		metaTemplatePaths    []string
		partialPaths         []string
		metaPaths            []string
		markdownContentPaths []string
		staticPaths          []string
	)

	logger := engine.Logger
	for _, filePath := range fileList.Files { // Check what type of file we have
		if strings.Contains(filePath, engine.PartialExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			partialPaths = append(partialPaths, filePath)
			logger.Debug("Identified as partial file", "path", filePath)
		} else if strings.Contains(filePath, engine.TemplateExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			templatePaths = append(templatePaths, filePath)
			logger.Debug("Identified as template file", "path", filePath)
		} else if strings.Contains(filePath, engine.MetaTemplateExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			metaTemplatePaths = append(metaTemplatePaths, filePath)
			logger.Debug("Identified as metatemplate file", "path", filePath)
		} else if path.Base(filePath) == engine.MetaFilename { // Making it easier to filter through them later and exclude them from staticPaths - they are not static files that should be copied to the outputDir
			metaPaths = append(metaPaths, filePath)
			logger.Debug("Identified as meta file", "path", filePath)
		} else if path.Base(filePath) == engine.MarkdownContentFilename { // Making it easier to filter through them later and exclude them from staticPaths - they are not static files that should be copied to the outputDir
			markdownContentPaths = append(markdownContentPaths, filePath)
			logger.Debug("Identified as markdown content file", "path", filePath)
		} else if slices.Contains(engine.ValuesFilePaths, filePath) { // Exclude values files from static files - they should not be copied to the outputDir
			logger.Debug("Identified as values file", "path", filePath)
		} else {
			staticPaths = append(staticPaths, filePath)
			logger.Debug("Identified as static file", "path", filePath)
		}
	}

	return templatePaths, metaTemplatePaths, partialPaths, metaPaths, markdownContentPaths, staticPaths
}
