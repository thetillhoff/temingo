package temingo

import "text/template"

// templateFuncMap returns the functions available to templates. It takes the
// engine because some functions - sri - need engine-scoped state.
func templateFuncMap(engine *Engine) template.FuncMap {
	return template.FuncMap{
		"concat":                 tmpl_concat,
		"includeWithIndentation": tmpl_indent,
		"capitalize":             tmpl_capitalize,
		"reverse":                tmpl_reverse,
		"sortBy":                 tmpl_sortBy,
		"filterBy":               tmpl_filterBy,
		"sri":                    engine.tmplSRI,
	}
}
