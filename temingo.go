package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/imdario/mergo"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var debug bool

func contains(list []string, search string) int {
	for index, x := range list {
		if x == search {
			return index
		}
	}
	return -1
}

func getTemplates(fromPath, extension string, exclusions []string) [][]string {
	var templates [][]string

	dirContents, err := ioutil.ReadDir(fromPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range dirContents {
		if !(entry.Name()[:1] == ".") { // ignore hidden files/folders
			entryPath := path.Join(fromPath, entry.Name())
			if fromPath == "." { // path.Join adds this to the filename directly ... which has to be prevented here
				entryPath = entry.Name()
			}
			if entry.IsDir() && (contains(exclusions, entryPath) == -1) {
				templates = append(templates, getTemplates(entryPath, extension, exclusions)...)
			} else if strings.HasSuffix(entry.Name(), extension) {
				if err != nil {
					log.Fatal(err)
				}
				fileContent, err := ioutil.ReadFile(entryPath)
				if err != nil {
					log.Fatal(err)
				}
				templates = append(templates, []string{entryPath, string(fileContent)})
			}
		}
	}

	return templates
}

func parseTemplateFiles(name string, baseTemplate string, partialTemplates [][]string) *template.Template {
	tpl := template.New(name)

	funcMap := sprig.HtmlFuncMap()

	extrafuncMap := template.FuncMap{
		"addPercentage": func(a string, b string) string {
			aInt, err := strconv.Atoi(a[:len(a)-1])
			if err != nil {
				log.Fatal(err)
			}
			bInt, err := strconv.Atoi(b[:len(b)-1])
			if err != nil {
				log.Fatal(err)
			}
			cInt := aInt + bInt
			return strconv.Itoa(cInt) + "%"
		},
		"include": func(name string, data map[string]interface{}) string {
			var buf strings.Builder
			err := tpl.ExecuteTemplate(&buf, name, data)
			if err != nil {
				log.Fatal(err)
			}
			result := buf.String()
			return result
		},
		"safehtml": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safecss": func(s string) template.CSS {
			return template.CSS(s)
		},
	}
	for k, v := range extrafuncMap {
		funcMap[k] = v
	}

	for index := range partialTemplates {
		partialTemplateContent := partialTemplates[index][1]
		_, err := tpl.Funcs(funcMap).Parse(partialTemplateContent)
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err := tpl.Funcs(funcMap).Parse(baseTemplate)
	if err != nil {
		log.Fatal(err)
	}
	return tpl
}

func writeTemplateToFile(filePath string, content []byte) error {
	dirPath := strings.TrimSuffix(filePath, path.Base(filePath))
	os.MkdirAll(dirPath, os.ModePerm)
	err := ioutil.WriteFile(filePath, content, os.ModePerm)
	return err
}

func cliFlags() ([]string, string, string, string, string, string, string) {
	var (
		valuesFilePaths    []string
		inputDirPath       string
		partialsDir        string
		outputDir          string
		templateExtension  string
		partialExtension   string
		generatedExtension string
	)

	flag.StringSliceVarP(&valuesFilePaths, "valuesfile", "f", []string{"values.yaml"}, "The path(s) to the values.yaml file(s)")
	flag.StringVarP(&inputDirPath, "inputDir", "i", ".", "The path to the template-file-directory")
	flag.StringVarP(&partialsDir, "partialsDir", "p", "partials", "The path to the partials-directory")
	flag.StringVarP(&outputDir, "outputDir", "o", "output", "The destination-path for the compiled templates")
	flag.StringVarP(&templateExtension, "templateExtension", "t", ".template", "The extension of the template files")
	flag.StringVar(&partialExtension, "partialExtension", ".partial", "The extension of the partial files") //TODO: not necessary, should be the same as templateExtension, since they are already distringuished by directory (or maybe force the extension to be *.partial.templateExtension)
	flag.StringVarP(&generatedExtension, "generatedExtension", "g", "", "The extension of the generated files")
	flag.BoolVarP(&debug, "debug", "d", false, "Setting this flag enables the debug mode")

	flag.Parse()

	// TODO: detect empty required flags & display help (with special flag or no flags at all)
	// if *valuesFilePath == "" || *inputDirPath == "" || {
	// 	fmt.Println("Usage:")
	// 	flag.PrintDefaults()
	// 	os.Exit(1)
	// }

	// TODO: add help flag
	// if flag.NFlag() == 0 { // if no flags are given display help
	// 	flag.PrintDefaults()
	// 	os.Exit(1)
	// }

	for i, valuesfilePath := range valuesFilePaths { // for each path stated
		valuesFilePaths[i] = path.Clean(valuesfilePath) // clean path
	}
	return valuesFilePaths, path.Clean(inputDirPath), path.Clean(partialsDir), path.Clean(outputDir), templateExtension, partialExtension, generatedExtension
}

func render(valuesFilePaths []string, inputDir string, partialsDir string, outputDir string, templateExtension string, partialExtension string, generatedExtension string) {
	// #####
	// START reading data file
	// #####
	if debug {
		log.Println("*** reading values file starts now ***")
	}

	var mappedValues map[string]interface{}
	for _, v := range valuesFilePaths {
		values, err := ioutil.ReadFile(v)
		if err != nil {
			log.Fatal(err)
		}
		var tempMappedValues map[string]interface{}
		yaml.Unmarshal([]byte(values), &tempMappedValues) // store yaml into map

		err = mergo.Merge(&mappedValues, tempMappedValues, mergo.WithOverride)
		if err != nil {
			log.Fatal(err)
		}
	}
	if debug {
		valuesYaml, err := yaml.Marshal(mappedValues)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(valuesYaml))
	}

	// #####
	// END reading data file
	// START collecting templates
	// #####
	if debug {
		log.Println("*** collecting templates starts now ***")
	}

	templates := getTemplates(inputDir, templateExtension, []string{path.Join(inputDir, partialsDir), path.Join(inputDir, outputDir)}) // get full html templates - with names
	partialTemplates := getTemplates(path.Join(inputDir, partialsDir), partialExtension, []string{path.Join(inputDir, outputDir)})     // get partial html templates - without names

	// #####
	// END collecting templates
	// START templating & output
	// #####
	if debug {
		log.Println("*** templating & output starts now ***")
	}

	outputBuffer := new(bytes.Buffer)
	for _, t := range templates {
		outputBuffer.Reset()
		tpl := parseTemplateFiles(t[0], t[1], partialTemplates)
		err := tpl.Execute(outputBuffer, mappedValues)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			os.MkdirAll(outputDir, 0700)
		}
		err = writeTemplateToFile(path.Join(outputDir, strings.TrimSuffix(tpl.Name(), templateExtension)+generatedExtension), outputBuffer.Bytes())
		if err != nil {
			log.Fatal(err)
		}
	}

	// #####
	// END templating & output
	// #####
}

func main() {
	// #####
	// START declaring variables
	// #####
	// no log.Println for debug here, because the flags have to be read first ;)

	valuesFilePaths, inputDirPath, partialsDir, outputDir, templateExtension, partialExtension, generatedExtension := cliFlags()
	// # example $> ./template -valuesfile values.yaml -inputDir ./ -partialsDir partials-html/ -templateExtension .html.template -generatedExtension .html

	if debug {
		log.Println("valuesFilePaths:", valuesFilePaths)
		log.Println("inputDirPath:", inputDirPath)
		log.Println("partialsDir:", partialsDir)
		log.Println("outputDir:", outputDir)
		log.Println("templateExtension:", templateExtension)
		log.Println("partialExtension:", partialExtension)
		log.Println("generatedExtension:", generatedExtension)
	}

	// #####
	// END declaring variables
	// START rendering
	// #####

	render(valuesFilePaths, inputDirPath, partialsDir, outputDir, templateExtension, partialExtension, generatedExtension)

	// #####
	// END rendering
	// #####
}
