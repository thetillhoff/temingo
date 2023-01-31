package fileIO

func ensureTrailingSlash(text string) string {

	if text[len(text)-1:] != "/" {
		return text + "/"
	} else {
		return text
	}
}
