package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

func loadYaml(filePath string) map[string]interface{} {
	var mappedObject map[string]interface{}
	values, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	yaml.Unmarshal([]byte(values), &mappedObject) // store yaml into map

	// valuesYaml, err := yaml.Marshal(mappedValues) // convert map to yaml/string
	return mappedObject
}
