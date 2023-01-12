package main

import (
	"log"
	"regexp"
)

var (
	debug bool
	watch bool

	valuesFilePaths         []string
	inputDir                string
	partialsDir             string
	outputDir               string
	staticDir               string
	templateExtension       string
	singleTemplateExtension string
	partialExtension        string
	temingoignoreFilePath   string

	listListObjects = make(map[string]map[string]interface{})

	pathValidator = "^[a-z0-9-_./]+$"
	rexp          = regexp.MustCompile(pathValidator)
)

type Breadcrumb struct {
	Name, Path interface{}
}

func main() {
	// #####
	// START declaring variables
	// #####

	// no log.Println for debug before this, because the flags have to be read first ;)
	readCliFlags()
	// # example $> ./template -valuesfile values.yaml -inputDir ./ -partialsDir partials-html/ -templateExtension .html.template -generatedExtension .html

	if debug {
		log.Println("valuesFilePaths:", valuesFilePaths)
		log.Println("inputDir:", inputDir)
		log.Println("partialsDir:", partialsDir)
		log.Println("outputDir:", outputDir)
		log.Println("templateExtension:", templateExtension)
		log.Println("singleTemplateExtension:", singleTemplateExtension)
		log.Println("partialExtension:", partialExtension)
		log.Println("temingoignoreFilePath:", temingoignoreFilePath)
		log.Println("staticDir:", staticDir)
		log.Println("watch:", watch)
	}

	// #####
	// END declaring variables
	// START rendering
	// #####

	if !watch { // if not watching
		rebuildOutput() // delete old contents of output-folder & copy static contents & render templates once
	} else { // else (== if watching)
		watchAll() // start to watch
	}

	// #####
	// END rendering
	// #####
}
