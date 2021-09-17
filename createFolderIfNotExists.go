package main

import "os"

func createFolderIfNotExists(path string) {
	os.MkdirAll(path, os.ModePerm)
}
