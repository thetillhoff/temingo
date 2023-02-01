package fileIO

// Returns the string with a trailing slash. If there already is one, it is returned as is
func ensureTrailingSlash(text string) string {

	if text[len(text)-1:] != "/" {
		return text + "/"
	} else {
		return text
	}
}
