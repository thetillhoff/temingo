package temingo

import "strings"

func tmpl_indent(indentation int, content string) string {
	indentationString := strings.Repeat(" ", indentation)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = indentationString + line
	}
	content = strings.Join(lines, "\n")
	return content
}
