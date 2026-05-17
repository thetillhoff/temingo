package temingo

import "text/template"

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"concat":                 tmpl_concat,
		"includeWithIndentation": tmpl_indent,
		"capitalize":             tmpl_capitalize,
		"reverse":                tmpl_reverse,
		"sortBy":                 tmpl_sortBy,
	}
}
