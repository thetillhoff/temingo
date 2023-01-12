package main

import (
	"log"
	"strings"
)

func createBreadcrumbs(path string) []Breadcrumb {
	if debug {
		log.Println("Creating breadcrumbs for '" + path + "'.")
	}
	breadcrumbs := []Breadcrumb{}
	currentPath := ""
	dirNames := strings.Split(path, "/")
	for ok := true; ok; ok = (len(dirNames) > 1) { // last one is not considered, so no self-reference occurs
		currentPath = currentPath + "/" + dirNames[0]
		breadcrumb := Breadcrumb{dirNames[0], currentPath}
		breadcrumbs = append(breadcrumbs, breadcrumb)
		dirNames = dirNames[1:] // remove first one, as it is now added to 'currentPath'
	}

	return breadcrumbs
}
