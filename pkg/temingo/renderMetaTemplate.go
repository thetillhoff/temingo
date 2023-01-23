package temingo

import (
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
)

func (engine *Engine) renderMetaTemplate(metaTemplatePath string, metaTemplateContent string, componentFiles map[string]string) (map[string][]byte, error) {
	var (
		err                    error
		files                  []fs.DirEntry
		metaTemplateRenderPath string
		renderedMetaTemplate   []byte
		renderedMetaTemplates  map[string][]byte
	)

	files, err = os.ReadDir(path.Dir(path.Join(engine.InputDir, metaTemplatePath))) // Get all child-elements of folder
	if err != nil {
		return renderedMetaTemplates, err
	}

	for _, f := range files { // For each child-element of folder
		if f.IsDir() { // Only for folders
			if engine.Verbose {
				log.Println("Found folder in metatemplatefolder", path.Join(engine.InputDir, metaTemplatePath, f.Name()))
			}
			if _, err = os.Stat(path.Join(engine.InputDir, metaTemplatePath, f.Name(), "meta.yaml")); !os.IsNotExist(err) { // Check if meta.yaml exists
				if engine.Verbose {
					log.Println("Found meta.yaml in", path.Join(engine.InputDir, metaTemplatePath, f.Name(), "meta.yaml"))
				}
				metaTemplateRenderPath = strings.ReplaceAll(path.Join(path.Dir(metaTemplatePath), f.Name(), path.Base(metaTemplatePath)), engine.MetaTemplateExtension, "")
				renderedMetaTemplate, err = engine.renderTemplate(metaTemplateRenderPath, metaTemplateContent, componentFiles) // By rendering as early as possible, related errors are also thrown very early. In this case, even before any filesystem changes are made.
				if err != nil {
					return renderedMetaTemplates, err
				}

				renderedMetaTemplates[metaTemplateRenderPath] = renderedMetaTemplate
			}
		}
	}

	return renderedMetaTemplates, nil

}
