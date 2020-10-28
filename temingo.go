package main

import (
	"bytes"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
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
	funcMap := template.FuncMap{
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
	}
	tpl := template.New(name)
	_, err := tpl.Funcs(funcMap).Parse(baseTemplate)
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

func writeTemplateToFile(filePath string, content []byte) error {
	dirPath := strings.TrimSuffix(filePath, path.Base(filePath))
	os.MkdirAll(dirPath, os.ModePerm)
	err := ioutil.WriteFile(filePath, content, os.ModePerm)
	return err
}

func cliFlags() (string, string, string, string, string, string, string) {
	valuesFilePath := flag.String("values", "", "the path to the values.yaml file")
	inputDirPath := flag.String("inputDir", "", "the path to the partials")
	partialsDir := flag.String("partialsDir", "", "the path to the partials")
	outputDir := flag.String("outputDir", "", "the path for the compiled templates")
	templateFileExtension := flag.String("templateFileExtension", "", "the extension of the template files")
	partialFileExtension := flag.String("partialFileExtension", "", "the extension of the partial files")
	generatedFileExtension := flag.String("generatedFileExtension", "", "the extension of the generated files")
	debugflag := flag.Bool("debug", false, "enable debug mode")

	flag.Parse()

	debug = *debugflag
	return path.Clean(*valuesFilePath), path.Clean(*inputDirPath), path.Clean(*partialsDir), path.Clean(*outputDir), *templateFileExtension, *partialFileExtension, *generatedFileExtension
}

func render(valuesFilePath string, inputDir string, partialsDir string, outputDir string, templateFileExtension string, partialFileExtension string, generatedFileExtension string) {
	// #####
	// START reading data file
	// #####
	if debug {
		log.Println("*** reading data file starts now ***")
	}

	values, err := ioutil.ReadFile(valuesFilePath)
	if err != nil {
		log.Fatal(err)
	}
	var mappedValues map[string]interface{}
	yaml.Unmarshal([]byte(values), &mappedValues) // store yaml into map

	// #####
	// END reading data file
	// START collecting templates
	// #####
	if debug {
		log.Println("*** collecting templates starts now ***")
	}

	templates := getTemplates(inputDir, templateFileExtension, []string{path.Join(inputDir, partialsDir), path.Join(inputDir, outputDir)}) // get full html templates - with names
	partialTemplates := getTemplates(path.Join(inputDir, partialsDir), partialFileExtension, []string{path.Join(inputDir, outputDir)})     // get partial html templates - without names

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
		err = tpl.Execute(outputBuffer, mappedValues)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			os.MkdirAll(outputDir, 0700)
		}
		err := writeTemplateToFile(path.Join(outputDir, strings.TrimSuffix(tpl.Name(), templateFileExtension)+generatedFileExtension), outputBuffer.Bytes())
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

	valuesFilePath, inputDirPath, partialsDir, outputDir, templateFileExtension, partialFileExtension, generatedFileExtension := cliFlags()
	// # example $> ./template -values values.yaml -inputDir ./ -partialsDir partials-html/ -templateFileExtension .html.template -generatedFileExtension .html

	if debug {
		log.Println("valuesFilePath:", valuesFilePath)
		log.Println("inputDirPath:", inputDirPath)
		log.Println("partialsDir:", partialsDir)
		log.Println("outputDir:", outputDir)
		log.Println("templateFileExtension:", templateFileExtension)
		log.Println("partialFileExtension:", partialFileExtension)
		log.Println("generatedFileExtension:", generatedFileExtension)
	}

	// #####
	// END declaring variables
	// START rendering
	// #####

	render(valuesFilePath, inputDirPath, partialsDir, outputDir, templateFileExtension, partialFileExtension, generatedFileExtension)

	// #####
	// END rendering
	// #####
}
