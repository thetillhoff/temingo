package temingo

import (
	"log"

	prettifyhtml "github.com/thetillhoff/temingo/pkg/prettifyHTML"
)

func (engine Engine) beautify(content []byte, ext string) []byte {

	switch ext {
	// TODO
	case ".html":
		if engine.Verbose {
			log.Println("beautified", ext)
		}
		return []byte(prettifyhtml.Format(string(content))) // Meh about the conversions
	default:
		return content
	}
}
