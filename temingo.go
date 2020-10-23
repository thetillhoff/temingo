package main

import (
	"bytes"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

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

func getTemplates(path, extension string, exclusions []string) [][]string {
	var templates [][]string

	dirContents, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range dirContents {
		if !(entry.Name()[:1] == ".") { // ignore hidden files/folders
			if entry.IsDir() && (contains(exclusions, path+entry.Name()) == -1) {
				templates = append(templates, getTemplates(path+entry.Name()+"/", extension, exclusions)...)
			} else if strings.HasSuffix(entry.Name(), extension) {
				if err != nil {
					log.Fatal(err)
				}
				fileContent, err := ioutil.ReadFile(path + "/" + entry.Name())
				if err != nil {
					log.Fatal(err)
				}
				templates = append(templates, []string{entry.Name(), string(fileContent)})
			}
		}
	}

	return templates
}

func parseTemplateFiles(name string, baseTemplate string, partialTemplates [][]string) *template.Template {
	funcMap := template.FuncMap{
		"incPercentage": func(a string, b string) string {
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
	}
	tpl := template.New(name)
	_, err := tpl.Parse(baseTemplate)
	if err != nil {
		log.Fatal(err)
	}
	for index := range partialTemplates {
		partialTemplateContent := partialTemplates[index][1]
		_, err := tpl.Funcs(funcMap).Parse(partialTemplateContent)
		if err != nil {
			log.Fatal(err)
		}
	}
	return tpl
}

func writeTemplateToFile(path string, content []byte) error {
	return ioutil.WriteFile(path, content, 0644)
}

func cliFlags() (string, string, string, string, string, string) {
	valuesFilePath := flag.String("values", "", "the path to the values.yaml file")
	inputDirPath := flag.String("inputDir", "", "the path to the partials")
	partialsDir := flag.String("partialsDir", "", "the path to the partials")
	outputDir := flag.String("outputDir", "", "the path for the compiled templates")
	templateFileExtension := flag.String("templateFileExtension", "", "the extension of the template files")
	generatedFileExtension := flag.String("generatedFileExtension", "", "the extension of the generated files")
	debugflag := flag.Bool("debug", false, "enable debug mode")

	flag.Parse()

	debug = *debugflag
	return *valuesFilePath, *inputDirPath, *partialsDir, *outputDir, *templateFileExtension, *generatedFileExtension
}

func main() {
	// #####
	// START variables
	// #####

	valuesFilePath, inputDirPath, partialsDir, outputDir, templateFileExtension, generatedFileExtension := cliFlags()
	// # example $> ./template -values values.yaml -inputDir ./ -partialsDir partials-html/ -templateFileExtension .html.template -generatedFileExtension .html

	if debug {
		log.Println("valuesFilePath:", valuesFilePath)
		log.Println("inputDirPath:", inputDirPath)
		log.Println("partialsDir:", partialsDir)
		log.Println("outputDir:", outputDir)
		log.Println("templateFileExtension:", templateFileExtension)
		log.Println("generatedFileExtension:", generatedFileExtension)
	}

	// #####
	// END variables
	// START read data file
	// #####

	if valuesFilePath[:1] != "/" && valuesFilePath[:2] != "./" && valuesFilePath[:3] != "../" { // if valuesFilePath is no absolute path and no full relative path
		valuesFilePath = "./" + valuesFilePath // add current folder as relative path
	}
	values, err := ioutil.ReadFile(valuesFilePath)
	if err != nil {
		log.Fatal(err)
	}
	var mappedValues map[string]interface{}
	yaml.Unmarshal([]byte(values), &mappedValues) // store yaml into map

	// #####
	// END read data file
	// START templating & output
	// #####

	templates := getTemplates(inputDirPath, templateFileExtension, []string{inputDirPath + partialsDir}) // get full html templates - with names
	partialTemplates := getTemplates(inputDirPath+partialsDir, templateFileExtension, []string{})        // get partial html templates - without names

	outputBuffer := new(bytes.Buffer)
	for _, t := range templates {
		outputBuffer.Reset()
		tpl := parseTemplateFiles(t[0], t[1], partialTemplates)
		err = tpl.Execute(outputBuffer, mappedValues)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			os.MkdirAll(outputDir, 0700)
		}
		err := writeTemplateToFile(outputDir+"/"+strings.TrimSuffix(tpl.Name(), templateFileExtension)+generatedFileExtension, outputBuffer.Bytes())
		if err != nil {
			log.Fatal(err)
		}
	}

	// #####
	// END templating & output
	// #####
}
