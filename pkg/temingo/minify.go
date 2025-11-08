package temingo

func (engine Engine) minify(content []byte, ext string) []byte {
	switch ext {
	// TODO
	default:
		engine.Logger.Warn("Minification not implemented for extension", "extension", ext)
		return content
	}
}
