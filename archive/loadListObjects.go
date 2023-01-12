package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

func loadListObjects(listPath string) map[string]interface{} {
	if debug {
		log.Println("*** Loading list objects from '" + listPath + "' ... ***")
	}
	contents, err := ioutil.ReadDir(path.Join(path.Clean("."), path.Clean(listPath)))
	if err != nil {
		log.Fatalln(err)
	}
	mappedObjects := make(map[string]interface{})
	for _, element := range contents {
		elementPath := path.Join(listPath, element.Name()) // f.e. list/element1 for folders
		indexPath := path.Join(elementPath, "index.yaml")  // f.e. list/element1/index.yaml
		if _, err := os.Stat(indexPath); err == nil {      // if list/element1/index.yaml exists
			if !rexp.MatchString(indexPath) { // if path is not good for urls
				log.Fatalln("The path '" + indexPath + "' for the list object must validate against the regular expression '" + pathValidator + "'.")
			}
			tempMappedObject := loadYaml(indexPath)      // f.e. list/element1/index.yaml
			tempMappedObject["Path"] = "/" + elementPath // will become /[.../]list/element1 (or actually /[.../]list/element1/index.html)
			mappedObjects[elementPath] = tempMappedObject
			if debug {
				log.Println("Loaded object from '" + indexPath + "' ...")
			}
		}
	}

	return mappedObjects
}
