package temingo

import "fmt"

func (engine *Engine) validateEngine() error {
	if engine.Beautify && engine.Minify {
		return fmt.Errorf("beautify and minify cannot both be enabled")
	}
	if engine.TemplateExtension == engine.MetaTemplateExtension {
		return fmt.Errorf("templateExtension and metaTemplateExtension must be different: %q", engine.TemplateExtension)
	}
	if engine.TemplateExtension == engine.PartialExtension {
		return fmt.Errorf("templateExtension and partialExtension must be different: %q", engine.TemplateExtension)
	}
	if engine.MetaTemplateExtension == engine.PartialExtension {
		return fmt.Errorf("metaTemplateExtension and partialExtension must be different: %q", engine.MetaTemplateExtension)
	}
	return nil
}
