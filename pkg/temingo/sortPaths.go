package temingo

import (
	"log"
	"strings"
)

func sortPaths(paths []string) ([]string, []string, []string, []string) {
	var (
		templatePaths     []string
		metaTemplatePaths []string
		componentPaths    []string
		staticPaths       []string
	)

	for _, filePath := range paths { // Check what type of file we have
		if strings.Contains(filePath, componentExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			componentPaths = append(componentPaths, filePath)
			if verbose {
				log.Println("Identified as component file:", filePath)
			}
		} else if strings.Contains(filePath, templateExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			templatePaths = append(templatePaths, filePath)
			if verbose {
				log.Println("Identified as template file:", filePath)
			}
		} else if strings.Contains(filePath, metaTemplateExtension) { // Multiple extensions are possible, so simply using path.Ext() is not enough (it only returns the last extension)
			metaTemplatePaths = append(metaTemplatePaths, filePath)
			if verbose {
				log.Println("Identified as metatemplate file:", filePath)
			}
		} else {
			staticPaths = append(staticPaths, filePath)
			if verbose {
				log.Println("Identified as static file:", filePath)
			}
		}
	}

	return templatePaths, metaTemplatePaths, componentPaths, staticPaths
}
