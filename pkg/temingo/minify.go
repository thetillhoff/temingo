package temingo

func (engine Engine) minify(content []byte, ext string) []byte {
	switch ext {
	default:
		engine.Logger.Warn("Minification not implemented for extension", "extension", ext)
		return content
	}
}

/*
General:
HTML and CSS is best minified together, because there are synergies.
- remove unused css classes
- split critical css and inline it
- autoconvert preload css
- separate stylesheets based on orientation (f.e. `<link media="(orientation: portrait)">`)
- separate stylesheets based on media queries (f.e. preferred-color-scheme, or `<link media="print">`)

For CSS:
- remove comments between `/* ... *\/`
- remove empty lines
- remove whitespace around selectors and properties, only leave brackets, commas and semicolons
- identify existing objects (no matter if hover, active, etc), and remove unused ones
- split critical css and inline it, while removing it from the original css file
  or use preload to indicate it's required even before the rendering starts
- separate stylesheets based on orientation (f.e. `<link media="(orientation: portrait)">`)
  optionally preload/fetch the other orientation as well.
- use devtools "Coverage" tool to identify unused / critical css
*/
