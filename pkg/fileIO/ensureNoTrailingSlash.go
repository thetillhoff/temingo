package fileIO

func ensureNoTrailingSlash(text string) string {

	if text[len(text)-1:] == "/" {
		return text[:len(text)-1]
	} else {
		return text
	}
}
