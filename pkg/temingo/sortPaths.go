package temingo

import (
	"log"
	"path"
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

	for _, filePath := range fileList.Files { // Check what type of file we have
		if strings.Contains(filePath, engine.PartialExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			partialPaths = append(partialPaths, filePath)
			if engine.Verbose {
				log.Println("Identified as partial file:", filePath)
			}
		} else if strings.Contains(filePath, engine.TemplateExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			templatePaths = append(templatePaths, filePath)
			if engine.Verbose {
				log.Println("Identified as template file:", filePath)
			}
		} else if strings.Contains(filePath, engine.MetaTemplateExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			metaTemplatePaths = append(metaTemplatePaths, filePath)
			if engine.Verbose {
				log.Println("Identified as metatemplate file:", filePath)
			}
		} else if path.Base(filePath) == engine.MetaFilename { // Making it easier to filter through them later and exclude them from staticPaths - they are not static files that should be copied to the outputDir
			metaPaths = append(metaPaths, filePath)
			if engine.Verbose {
				log.Println("Identified as meta file:", filePath)
			}
		} else if path.Base(filePath) == engine.MarkdownContentFilename { // Making it easier to filter through them later and exclude them from staticPaths - they are not static files that should be copied to the outputDir
			markdownContentPaths = append(markdownContentPaths, filePath)
			if engine.Verbose {
				log.Println("Identified as markdown content file:", filePath)
			}
		} else if engine.ValuesFilePath != "" && filePath == engine.ValuesFilePath { // Exclude values file from static files - it should not be copied to the outputDir
			if engine.Verbose {
				log.Println("Identified as values file:", filePath)
			}
		} else {
			staticPaths = append(staticPaths, filePath)
			if engine.Verbose {
				log.Println("Identified as static file:", filePath)
			}
		}
	}

	return templatePaths, metaTemplatePaths, partialPaths, metaPaths, markdownContentPaths, staticPaths
}
