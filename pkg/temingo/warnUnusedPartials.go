package temingo

import "regexp"

var templateCallRe = regexp.MustCompile(`\{\{-?\s*template\s+"([^"]+)"`)

// warnUnusedPartials logs a warning for each defined partial that is never referenced
// by any template or other partial content.
func (engine *Engine) warnUnusedPartials(partialFiles map[string]string, allTemplateContents []string) {
	referenced := map[string]bool{}
	for _, content := range allTemplateContents {
		for _, match := range templateCallRe.FindAllStringSubmatch(content, -1) {
			referenced[match[1]] = true
		}
	}
	// Also scan partial content itself so partials-calling-partials are counted
	for _, content := range partialFiles {
		for _, match := range templateCallRe.FindAllStringSubmatch(content, -1) {
			referenced[match[1]] = true
		}
	}

	for partialPath := range partialFiles {
		if !referenced[partialPath] {
			engine.Logger.Warn("Unused partial", "path", partialPath)
		}
	}
}
