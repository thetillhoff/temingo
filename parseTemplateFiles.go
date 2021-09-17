package main

import (
	"html/template"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/PuerkitoBio/purell"
	"github.com/imdario/mergo"
)

func parseTemplateFiles(name string, baseTemplate string, partialTemplates [][]string) *template.Template {
	tpl := template.New(name)

	funcMap := sprig.HtmlFuncMap()

	extrafuncMap := template.FuncMap{
		"addPercentage": func(a string, b string) string {
			aInt, err := strconv.Atoi(a[:len(a)-1])
			if err != nil {
				log.Fatalln(err)
			}
			bInt, err := strconv.Atoi(b[:len(b)-1])
			if err != nil {
				log.Fatalln(err)
			}
			cInt := aInt + bInt
			return strconv.Itoa(cInt) + "%"
		},
		"include": func(name string, data map[string]interface{}) string {
			var buf strings.Builder
			err := tpl.ExecuteTemplate(&buf, name, data)
			if err != nil {
				log.Fatalln(err)
			}
			result := buf.String()
			return result
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
		"list": func(listPaths ...string) map[string]interface{} {
			listObjects := make(map[string]interface{})
			if len(listPaths) == 0 { // If no path is provided
				listPaths = append(listPaths, filepath.Dir(name)) // Add the default path (folder containing the template)
			}
			for _, listPath := range listPaths {
				mergo.Merge(&listObjects, loadListObjects(listPath))
				listListObjects[listPath] = listObjects
			}
			return listObjects
		},
		"urlize": func(oldContent string) string {
			newContent, err := purell.NormalizeURLString(strings.ReplaceAll(oldContent, " ", "_"), purell.FlagsSafe)
			if err != nil {
				log.Fatalln(err)
			}
			newContent = strings.ToLower(newContent) // Also convert everything to lowercase. Arguable.
			if debug {
				log.Println("Urlized '" + oldContent + "' to '" + newContent + "'.")
			}
			return newContent
		},
		"capitalize": func(oldContent string) string {
			newContent := strings.Title(oldContent)
			if debug {
				log.Println("Capitalized '" + oldContent + "' to '" + newContent + "'.")
			}
			return newContent
		},
	}
	for k, v := range extrafuncMap {
		funcMap[k] = v
	}

	for index := range partialTemplates {
		partialTemplateContent := partialTemplates[index][1]
		_, err := tpl.Funcs(funcMap).Parse(partialTemplateContent)
		if err != nil {
			log.Fatalln(err)
		}
	}
	_, err := tpl.Funcs(funcMap).Parse(baseTemplate)
	if err != nil {
		log.Fatalln(err)
	}
	return tpl
}
