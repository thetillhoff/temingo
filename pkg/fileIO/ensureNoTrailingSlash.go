package fileIO

// Returns the string without a trailing slash. If there is none, it is returned as is
func ensureNoTrailingSlash(text string) string {

	if text[len(text)-1:] == "/" {
		return text[:len(text)-1]
	} else {
		return text
	}
}
