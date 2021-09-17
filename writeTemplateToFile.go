package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func writeTemplateToFile(filePath string, content []byte) error {
	dirPath := strings.TrimSuffix(filePath, path.Base(filePath))
	createFolderIfNotExists(dirPath)
	err := ioutil.WriteFile(filePath, content, os.ModePerm)
	return err
}
