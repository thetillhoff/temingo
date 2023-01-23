package temingo

import (
	"log"
	"path"
	"strings"

	"github.com/thetillhoff/temingo/pkg/fileIO"
)

func (engine *Engine) renderMetaTemplate(fileList fileIO.FileList, metaTemplatePath string, metaTemplateContent string, componentFiles map[string]string) (map[string][]byte, error) {
	var (
		err error
		// files                  []fs.DirEntry
		metaTemplateRenderPath string
		renderedMetaTemplate   []byte
		renderedMetaTemplates  = make(map[string][]byte)
	)
	for _, metaFilePath := range fileList.FilterByLevelAtPath(path.Dir(metaTemplatePath), 2).FilterByFileName(defaultMetaFileName).Files {
		if engine.Verbose {
			log.Println("Found metatemplate child at", metaFilePath)
		}

		metaTemplateRenderPath = strings.ReplaceAll(path.Join(path.Dir(metaFilePath), path.Base(metaTemplatePath)), engine.MetaTemplateExtension, "")
		renderedMetaTemplate, err = engine.renderTemplate(metaTemplateRenderPath, metaTemplateContent, componentFiles) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
		if err != nil {
			return renderedMetaTemplates, err
		}

		renderedMetaTemplates[metaTemplateRenderPath] = renderedMetaTemplate
	}

	return renderedMetaTemplates, nil

}
