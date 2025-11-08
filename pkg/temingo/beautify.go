package temingo

import (
	prettifyhtml "github.com/thetillhoff/temingo/pkg/prettifyHTML"
)

func (engine Engine) beautify(content []byte, ext string) []byte {
	logger := engine.Logger

	switch ext {
	// TODO add more
	case ".html":
		logger.Debug("Beautifying content", "extension", ext)
		return []byte(prettifyhtml.Format(string(content))) // Meh about the conversions
	default:
		return content
	}
}
