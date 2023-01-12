package main

import (
	"log"
	"os"
	"path"

	gitignore "github.com/sabhiram/go-gitignore"
)

func isExcluded(srcPath string, additionalExclusions []string) bool {
	var (
		ignore *gitignore.GitIgnore
	)

	srcPath = "/" + srcPath

	additionalExclusions = append(additionalExclusions, "/"+temingoignoreFilePath)      // always ignore the ignore file itself
	additionalExclusions = append(additionalExclusions, "/"+path.Join(outputDir, "**")) // always ignore the outputDir
	additionalExclusions = append(additionalExclusions, "/"+path.Join(staticDir, "**")) // always ignore the staticDir

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

	if ignore.MatchesPath((srcPath)) {
		if debug {
			log.Println("Exclusion triggered at '" + srcPath + "', specified internally.")
		}
		return true
	}

	return false
}
