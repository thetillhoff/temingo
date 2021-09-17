package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/otiai10/copy"
)

func rebuildOutput() {
	// #####
	// START Delete output-dir contents
	// #####

	if debug {
		log.Println("*** Deleting contents in output-dir ... ***")
	}

	dirContents, err := ioutil.ReadDir(outputDir)
	if err != nil {
		log.Fatalln(err)
	}
	for _, element := range dirContents {
		elementPath := path.Join(outputDir, element.Name())
		if debug {
			log.Println("Deleting output-dir content at: " + elementPath)
		}
		err = os.RemoveAll(elementPath)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// #####
	// END Delete output-dir contents
	// START Copy static-dir contents to output-dir
	// #####

	if debug {
		log.Println("*** Copying contents of static-dir to output-dir ... ***")
	}

	err = copy.Copy(staticDir, outputDir)
	if err != nil {
		log.Fatalln(err)
	}

	// #####
	// END Copy static-dir-contents to output-dir
	// START Copy other contents to output-dir
	// #####

	if debug {
		log.Println("*** Copying other contents to output-dir ... ***")
	}

	opt := copy.Options{
		Skip: func(src string) (bool, error) {
			skip := false
			if isExcluded(src, []string{path.Join("/", partialsDir), "**/*" + templateExtension, "**/index.yaml"}) || isExcludedByTemingoignore(src, []string{}) {
				skip = true
			}
			return skip, nil
		},
	}
	err = copy.Copy(inputDir, outputDir, opt)
	if err != nil {
		log.Fatalln(err)
	}

	// #####
	// END Copy other contents to output-dir
	// START Render templates
	// #####

	if debug {
		log.Println("*** Starting templating process ... ***")
	}

	render()
	log.Println("*** Successfully built contents. ***")

	// #####
	// END Render templates
	// #####
}
