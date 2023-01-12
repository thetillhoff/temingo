package main

import (
	"io/ioutil"
	"log"
	"path"
	"strings"
)

func getTemplates(fromPath string, extension string, additionalExclusions []string) [][]string {
	var templates [][]string

	dirContents, err := ioutil.ReadDir(fromPath)
	if err != nil {
		log.Fatalln(err)
	}
	for _, entry := range dirContents {
		if !(entry.Name()[:1] == ".") { // ignore hidden files/folders
			entryPath := path.Join(fromPath, entry.Name())
			if fromPath == "." { // path.Join adds this to the filename directly ... which has to be prevented here
				entryPath = entry.Name()
			}
			if !isExcluded(entryPath, additionalExclusions) { // Make all paths absolute from working-directory
				if entry.IsDir() {
					templates = append(templates, getTemplates(entryPath, extension, additionalExclusions)...)
				} else if strings.HasSuffix(entry.Name(), extension) {
					if !rexp.MatchString(entryPath) {
						log.Fatalln("The path '" + entryPath + "' doesn't validate against the regular expression '" + pathValidator + "'.")
					}
					fileContent, err := ioutil.ReadFile(entryPath)
					if err != nil {
						log.Fatalln(err)
					}
					templates = append(templates, []string{entryPath, string(fileContent)})
				}
			}
		}
	}

	return templates
}
