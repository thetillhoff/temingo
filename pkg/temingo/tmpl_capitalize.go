package temingo

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func tmpl_capitalize(s string) string {
	return cases.Title(language.English).String(s)
}
