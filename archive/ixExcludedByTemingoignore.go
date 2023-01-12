package main

import (
	"log"
	"os"

	gitignore "github.com/sabhiram/go-gitignore"
)

func isExcludedByTemingoignore(srcPath string, additionalExclusions []string) bool {
	var (
		ignore *gitignore.GitIgnore
	)

	srcPath = "/" + srcPath

	// If the temingoignorefile does not exist, only add manual additions
	if _, err := os.Stat(temingoignoreFilePath); os.IsNotExist(err) {
		ignore = gitignore.CompileIgnoreLines(additionalExclusions...)
	} else if err != nil {
		log.Fatalln(err)
	} else {
		ignore, err = gitignore.CompileIgnoreFileAndLines(temingoignoreFilePath, additionalExclusions...)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if ignore.MatchesPath(srcPath) {
		if debug {
			log.Println("Exclusion triggered at '" + srcPath + "', specified in '" + temingoignoreFilePath + "'.")
		}
		return true
	}

	return false
}
