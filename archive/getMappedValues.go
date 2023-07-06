package main

import (
	"log"

	"dario.cat/mergo"
)

func getMappedValues() map[string]interface{} {
	var mappedValues map[string]interface{}
	for _, v := range valuesFilePaths {
		tempMappedValues := loadYaml(v)

		err := mergo.Merge(&mappedValues, tempMappedValues, mergo.WithOverride)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return mappedValues
}
