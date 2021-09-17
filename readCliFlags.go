package main

import (
	"log"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

func readCliFlags() {
	var (
		info os.FileInfo
		err  error
	)

	flag.StringSliceVarP(&valuesFilePaths, "valuesfile", "f", []string{"values.yaml"}, "Sets the path(s) to the values-file(s).")
	flag.StringVarP(&inputDir, "inputDir", "i", ".", "Sets the path to the template-file-directory.")
	flag.StringVarP(&partialsDir, "partialsDir", "p", "partials", "Sets the path to the partials-directory.")
	flag.StringVarP(&outputDir, "outputDir", "o", "output", "Sets the destination-path for the compiled templates.")
	flag.StringVarP(&staticDir, "staticDir", "s", "static", "Sets the source-path for the static files.")
	flag.StringVarP(&templateExtension, "templateExtension", "t", ".template", "Sets the extension of the template files.")
	flag.StringVar(&singleTemplateExtension, "singleTemplateExtension", ".single.template", "Sets the extension of the single-view template files. Automatically excluded from normally loaded templates.")
	flag.StringVar(&partialExtension, "partialExtension", ".partial", "Sets the extension of the partial files.") //TODO: not necessary, should be the same as templateExtension, since they are already distringuished by directory -> Might be useful when "modularization" will be implemented
	flag.StringVar(&temingoignoreFilePath, "temingoignore", ".temingoignore", "Sets the path to the ignore file.")
	flag.BoolVarP(&watch, "watch", "w", false, "Watches the template-file-directory, partials-directory and values-files.")
	flag.BoolVarP(&debug, "debug", "d", false, "Enables the debug mode.")

	flag.Parse() // Actually read the configured cli-flags

	for i, valuesfilePath := range valuesFilePaths { // for each path stated
		valuesFilePaths[i] = path.Clean(valuesfilePath) // clean path
		info, err = os.Stat(valuesFilePaths[i])
		if os.IsNotExist(err) { // if path doesn't exist
			log.Fatalln("Values file does not exist: " + valuesFilePaths[i])
		} else if info.IsDir() { // if is not a directoy
			log.Fatalln("Values file is not a file (but a directory): " + valuesFilePaths[i])
		}
	}

	inputDir = path.Clean(inputDir)
	info, err = os.Stat(inputDir)
	if os.IsNotExist(err) { // if path doesn't exist
		log.Fatalln("Given input-directory does not exist: " + inputDir)
	} else if !info.IsDir() { // if is not a directory
		log.Fatalln("Given input-directory is not a directory: " + inputDir)
	}

	partialsDir = path.Clean(partialsDir)
	info, err = os.Stat(partialsDir)
	if os.IsNotExist(err) { // if path doesn't exist
		log.Fatalln("Given partial-files-directory does not exist: " + partialsDir)
	} else if !info.IsDir() { // if is not a directory
		log.Fatalln("Given partial-files-directory is not a directory: " + partialsDir)
	}

	outputDir = path.Clean(outputDir)
	info, err = os.Stat(outputDir)
	if os.IsNotExist(err) { // if path doesn't exist
		log.Fatalln("Given output-directory does not exist: " + outputDir)
	} else if !info.IsDir() { // if is not a directory
		log.Fatalln("Given output-directory is not a directory: " + outputDir)
	}

	staticDir = path.Clean(staticDir)
	info, err = os.Stat(staticDir)
	if os.IsNotExist(err) { // if path doesn't exist
		log.Fatalln("Given static-files-directory does not exist: " + staticDir)
	} else if !info.IsDir() { // if is not a directory
		log.Fatalln("Given static-files-directory is not a directory: " + staticDir)
	}
}
