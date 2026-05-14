package temingo

import "strings"

// warnHTTPLinks logs a warning for each http:// link found in the rendered content,
// excluding localhost and loopback addresses which are intentionally non-TLS.
func (engine *Engine) warnHTTPLinks(filePath string, content []byte) {
	s := string(content)
	offset := 0
	for {
		idx := strings.Index(s[offset:], "http://")
		if idx == -1 {
			break
		}
		abs := offset + idx
		// Extract the URL up to the next whitespace or quote for a readable log message
		end := abs + 7
		for end < len(s) && s[end] != '"' && s[end] != '\'' && s[end] != ' ' && s[end] != '\n' && s[end] != '\t' && s[end] != '<' {
			end++
		}
		url := s[abs:end]

		if !strings.HasPrefix(url, "http://localhost") && !strings.HasPrefix(url, "http://127.0.0.1") {
			engine.Logger.Warn("Insecure http:// link found", "file", filePath, "url", url)
		}
		offset = abs + 7
	}
}
