package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func render() {
	// #####
	// START reading value files
	// #####
	if debug {
		log.Println("*** Reading values file(s) ... ***")
	}
	mappedValues := getMappedValues()
	if debug {
		valuesYaml, err := yaml.Marshal(mappedValues)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("*** General values-object: ***\n" + string(valuesYaml))
	}

	// #####
	// END reading value files
	// START normal templating
	// #####

	templates := getTemplates(inputDir, templateExtension, []string{"**/*" + singleTemplateExtension}) // get full html templates - with names
	partialTemplates := getTemplates(partialsDir, partialExtension, []string{})                        // get partial html templates - without names

	for _, template := range templates {
		outputFilePath := path.Join(outputDir, strings.TrimSuffix(template[0], templateExtension))
		if debug {
			log.Println("Writing output file '" + outputFilePath + "' ...")
		}
		runTemplate(mappedValues, template[0], template[1], partialTemplates, outputFilePath)
	}

	// #####
	// END normal templating
	// START single-view templating
	// #####

	// identify & collect single-view templates via their extension
	singleTemplates := getTemplates(inputDir, singleTemplateExtension, []string{
		path.Join(inputDir, partialsDir, "**"),
		path.Join(inputDir, outputDir, "**"),
	}) // get full html templates - with names

	// for each of the single-view templates
	for _, template := range singleTemplates {
		templateName := template[0]
		template := template[1]
		// search all configurations

		dirContents, err := ioutil.ReadDir(filepath.Dir(templateName))
		if err != nil {
			log.Fatalln(err)
		}

		itemValues := make(map[string]interface{})

		// Read item-specific values, so they are available independent of the items way of the configuration
		for _, dirEntry := range dirContents {
			if dirEntry.IsDir() {
				if _, err := os.Stat(path.Join(filepath.Dir(templateName), dirEntry.Name(), "index.yaml")); err == nil { // if the dirEntry-folder contains an "index.yaml"
					itemValues[path.Join(filepath.Dir(templateName), dirEntry.Name())] = loadYaml(path.Join(filepath.Dir(templateName), dirEntry.Name(), "index.yaml"))
				}
			}
		}

		for itemPath, itemValue := range itemValues {
			// load corresponding additional values into mappedValues["Item"]
			extendedMappedValues := mappedValues
			itemPath = strings.TrimSuffix(itemPath, filepath.Ext(itemPath))
			fileName := strings.TrimSuffix(filepath.Base(templateName), singleTemplateExtension)
			extendedMappedValues["ItemPath"] = "/" + itemPath
			extendedMappedValues["Item"] = itemValue
			outputFilePath := path.Join(outputDir, itemPath, fileName)
			if debug {
				log.Println("Writing single-view output from '" + itemPath + "*' to '" + outputFilePath + "' ...") // itemPath is incomplete; either its a yaml-file or a folder containing an index.yaml -> Therefore it has the '*' behind it.
			}
			runTemplate(extendedMappedValues, templateName, template, partialTemplates, outputFilePath)
		}
	}

	// #####
	// END single-view templating
	// #####
}
