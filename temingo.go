package main

import (
	"bytes"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/imdario/mergo"
	"github.com/otiai10/copy"
	"github.com/rjeczalik/notify"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	debug bool
	watch bool

	valuesFilePaths    []string
	inputDir           string
	partialsDir        string
	outputDir          string
	staticDir          string
	templateExtension  string
	partialExtension   string
	generatedExtension string

	listListObjects = make(map[string]map[string]interface{})
)

func contains(list []string, search string) int {
	for index, x := range list {
		if x == search {
			return index
		}
	}
	return -1
}

func getTemplates(fromPath string, extension string, exclusions []string) [][]string {
	var templates [][]string

	dirContents, err := ioutil.ReadDir(fromPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range dirContents {
		if !(entry.Name()[:1] == ".") { // ignore hidden files/folders
			entryPath := path.Join(fromPath, entry.Name())
			if fromPath == "." { // path.Join adds this to the filename directly ... which has to be prevented here
				entryPath = entry.Name()
			}
			if entry.IsDir() && (contains(exclusions, entryPath) == -1) {
				templates = append(templates, getTemplates(entryPath, extension, exclusions)...)
			} else if strings.HasSuffix(entry.Name(), extension) {
				if err != nil {
					log.Fatal(err)
				}
				fileContent, err := ioutil.ReadFile(entryPath)
				if err != nil {
					log.Fatal(err)
				}
				templates = append(templates, []string{entryPath, string(fileContent)})
			}
		}
	}

	return templates
}

func parseTemplateFiles(name string, baseTemplate string, partialTemplates [][]string) *template.Template {
	tpl := template.New(name)

	funcMap := sprig.HtmlFuncMap()

	extrafuncMap := template.FuncMap{
		"addPercentage": func(a string, b string) string {
			aInt, err := strconv.Atoi(a[:len(a)-1])
			if err != nil {
				log.Fatal(err)
			}
			bInt, err := strconv.Atoi(b[:len(b)-1])
			if err != nil {
				log.Fatal(err)
			}
			cInt := aInt + bInt
			return strconv.Itoa(cInt) + "%"
		},
		"include": func(name string, data map[string]interface{}) string {
			var buf strings.Builder
			err := tpl.ExecuteTemplate(&buf, name, data)
			if err != nil {
				log.Fatal(err)
			}
			result := buf.String()
			return result
		},
		"safehtml": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safecss": func(s string) template.CSS {
			return template.CSS(s)
		},
		"list": func(listPath string) map[string]interface{} {
			listObjects := loadListObjects(listPath)
			listListObjects[listPath] = listObjects
			return listObjects
		},
	}
	for k, v := range extrafuncMap {
		funcMap[k] = v
	}

	for index := range partialTemplates {
		partialTemplateContent := partialTemplates[index][1]
		_, err := tpl.Funcs(funcMap).Parse(partialTemplateContent)
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err := tpl.Funcs(funcMap).Parse(baseTemplate)
	if err != nil {
		log.Fatal(err)
	}
	return tpl
}

func writeTemplateToFile(filePath string, content []byte) error {
	dirPath := strings.TrimSuffix(filePath, path.Base(filePath))
	os.MkdirAll(dirPath, os.ModePerm)
	err := ioutil.WriteFile(filePath, content, os.ModePerm)
	return err
}

func readCliFlags() {
	var (
		info fs.FileInfo
		err  error
	)

	flag.StringSliceVarP(&valuesFilePaths, "valuesfile", "f", []string{"values.yaml"}, "Sets the path(s) to the values-file(s)")
	flag.StringVarP(&inputDir, "inputDir", "i", ".", "Sets the path to the template-file-directory")
	flag.StringVarP(&partialsDir, "partialsDir", "p", "partials", "Sets the path to the partials-directory")
	flag.StringVarP(&outputDir, "outputDir", "o", "output", "Sets the destination-path for the compiled templates")
	flag.StringVarP(&staticDir, "staticDir", "s", "static", "Sets the source-path for the static files")
	flag.StringVarP(&templateExtension, "templateExtension", "t", ".template", "Sets the extension of the template files")
	flag.StringVar(&partialExtension, "partialExtension", ".partial", "Sets the extension of the partial files") //TODO: not necessary, should be the same as templateExtension, since they are already distringuished by directory -> Might be useful when "modularization" will be implemented
	flag.StringVarP(&generatedExtension, "generatedExtension", "g", "", "Sets the extension of the generated files")
	flag.BoolVarP(&watch, "watch", "w", false, "Watches the template-file-directory, partials-directory and values-files")
	flag.BoolVarP(&debug, "debug", "d", false, "Enables the debug mode")

	flag.Parse() // Actually read the configured cli-flags

	for i, valuesfilePath := range valuesFilePaths { // for each path stated
		valuesFilePaths[i] = path.Clean(valuesfilePath) // clean path
		info, err = os.Stat(valuesFilePaths[i])
		if os.IsNotExist(err) { // if path doesn't exist
			log.Fatal("Values file does not exist: " + valuesFilePaths[i])
		} else if info.IsDir() { // if is not a directoy
			log.Fatal("Values file is not a file (but a directory): " + valuesFilePaths[i])
		}
	}

	inputDir = path.Clean(inputDir)
	info, err = os.Stat(inputDir)
	if os.IsNotExist(err) { // if path doesn't exist
		log.Fatal("Given input-directory does not exist: " + inputDir)
	} else if !info.IsDir() { // if is not a directory
		log.Fatal("Given input-directory is not a directory: " + inputDir)
	}

	partialsDir = path.Clean(partialsDir)
	info, err = os.Stat(partialsDir)
	if os.IsNotExist(err) { // if path doesn't exist
		log.Fatal("Given partial-files-directory does not exist: " + partialsDir)
	} else if !info.IsDir() { // if is not a directory
		log.Fatal("Given partial-files-directory is not a directory: " + partialsDir)
	}

	outputDir = path.Clean(outputDir)
	info, err = os.Stat(outputDir)
	if os.IsNotExist(err) { // if path doesn't exist
		log.Fatal("Given output-directory does not exist: " + outputDir)
	} else if !info.IsDir() { // if is not a directory
		log.Fatal("Given output-directory is not a directory: " + outputDir)
	}

	staticDir = path.Clean(staticDir)
	info, err = os.Stat(staticDir)
	if os.IsNotExist(err) { // if path doesn't exist
		log.Fatal("Given static-files-directory does not exist: " + staticDir)
	} else if !info.IsDir() { // if is not a directory
		log.Fatal("Given static-files-directory is not a directory: " + staticDir)
	}
}

func render() {
	// #####
	// START reading data file
	// #####
	if debug {
		log.Println("*** Reading values file(s) ... ***")
	}

	var mappedValues map[string]interface{}
	for _, v := range valuesFilePaths {
		tempMappedValues := loadYaml(v)

		err := mergo.Merge(&mappedValues, tempMappedValues, mergo.WithOverride)
		if err != nil {
			log.Fatal(err)
		}
	}
	if debug {
		valuesYaml, err := yaml.Marshal(mappedValues)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("*** values-object: ***\n" + string(valuesYaml))
	}

	// #####
	// END reading data file
	// START collecting templates
	// #####
	if debug {
		log.Println("*** Collecting templates ... ***")
	}

	templates := getTemplates(inputDir, templateExtension, []string{path.Join(inputDir, partialsDir), path.Join(inputDir, outputDir)}) // get full html templates - with names
	partialTemplates := getTemplates(path.Join(inputDir, partialsDir), partialExtension, []string{path.Join(inputDir, outputDir)})     // get partial html templates - without names

	// #####
	// END collecting templates
	// START templating & output
	// #####
	if debug {
		log.Println("*** Templating & writing output files ... ***")
	}

	outputBuffer := new(bytes.Buffer)
	for _, t := range templates {
		outputBuffer.Reset()
		tpl := parseTemplateFiles(t[0], t[1], partialTemplates)
		err := tpl.Execute(outputBuffer, mappedValues)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			os.MkdirAll(outputDir, 0700)
		}
		outputFilePath := path.Join(outputDir, strings.TrimSuffix(tpl.Name(), templateExtension)+generatedExtension)
		if debug {
			log.Println("Writing output file '" + outputFilePath + "' ...")
		}
		err = writeTemplateToFile(outputFilePath, outputBuffer.Bytes())
		if err != nil {
			log.Fatal(err)
		}
	}

	// #####
	// END templating & output
	// START list & single view
	// #####

	for listPath, listObject := range listListObjects { // for each list
		sourcePath := path.Join(outputDir, listPath, "index.html.single")
		if _, err := os.Stat(sourcePath); err == nil { // check if single-view template file exists
			for key := range listObject { // for each object
				if strings.HasSuffix(key, ".yaml") { // for list-files
					destinationPath := path.Join(outputDir, strings.TrimSuffix(key, filepath.Ext(key)), "index.html")
					os.Mkdir(filepath.Base(destinationPath), os.ModePerm)
					if debug {
						log.Println("copying list-file: " + sourcePath + " to " + destinationPath)
					}
					copy.Copy(sourcePath, destinationPath)
				} else { // for list-folders
					destinationPath := path.Join(outputDir, key, "index.html")
					if debug {
						log.Println("copying list-folder: " + sourcePath + " to " + destinationPath)
					}
					copy.Copy(sourcePath, destinationPath)
				}
			}
			os.Remove(sourcePath) // delete single-view template file
		}
	}

	// #####
	// END list & single view
	// #####
}

func watchAll() {
	if debug {
		log.Println("*** Starting to watch for filesystem changes ... ***")
	}

	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1)
	// Multiple calls for the channel only expands the event sent, not overwrites it (see https://pkg.go.dev/github.com/rjeczalik/notify?utm_source=godoc#Watch)
	// Set up a watchpoint listening for events within a directory tree rooted at current working directory.
	// Events taken from https://pkg.go.dev/github.com/rjeczalik/notify?utm_source=godoc#pkg-constants
	if err := notify.Watch(inputDir+"/...", c, notify.Create, notify.Remove, notify.Write, notify.Rename); err != nil { // watch the input-files-directory recursively (for all events)
		log.Fatal(err)
	}
	if err := notify.Watch(inputDir+"/...", c, notify.Create, notify.Remove, notify.Write, notify.Rename); err != nil { // watch the partials-files-directory recursively (for all events)
		log.Fatal(err)
	}
	for _, valuesFile := range valuesFilePaths { // for each valuesfilepath
		if err := notify.Watch(valuesFile, c, notify.Write); err != nil { // watch the path (only for writes/changes)
			log.Print(err) // Don't fail/crash, but continue on next save
		}
	}

	// Clean up watchpoint associated with c. If Stop was not called upon
	// return the channel would be leaked as notify holds the only reference
	// to it and does not release it on its own.
	defer notify.Stop(c)

	for { // while true
		// Block until an event is received.
		ei := <-c

		if debug {
			log.Println("filesystem-change notification received:", ei)
		}

		rebuildOutput()
	}
}

func rebuildOutput() {
	// #####
	// START Delete output-dir contents
	// #####

	if debug {
		log.Println("*** Deleting contents in output-dir ... ***")
	}

	dirContents, err := ioutil.ReadDir(outputDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, element := range dirContents {
		elementPath := path.Join(outputDir, element.Name())
		if debug {
			log.Println("output-dir: " + elementPath)
		}
		err = os.RemoveAll(elementPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	// #####
	// END Delete output-dir contents
	// START Copy static-dir contents to output-dir
	// #####

	if debug {
		log.Println("*** Copying contents of static-dir to output-dir ... ***")
	}

	err = copy.Copy(staticDir, outputDir)
	if err != nil {
		log.Fatal(err)
	}

	// #####
	// END Copy static-dir-contents to output-dir
	// START Render templates
	// #####

	render()

	if debug { //TODO delete this output, and only display count
		log.Println(listListObjects)
	}

	// #####
	// END Render templates
	// #####
}

func loadYaml(filePath string) map[string]interface{} {
	var mappedObject map[string]interface{}
	values, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	yaml.Unmarshal([]byte(values), &mappedObject) // store yaml into map

	// valuesYaml, err := yaml.Marshal(mappedValues) // convert map to yaml/string
	return mappedObject
}

func loadListObjects(listPath string) map[string]interface{} {
	if debug {
		log.Print("*** Loading list objects from '" + listPath + "' ... ***")
	}
	contents, err := ioutil.ReadDir(path.Join(path.Clean("."), path.Clean(listPath)))
	if err != nil {
		log.Fatal(err)
	}
	mappedObjects := make(map[string]interface{})
	for _, element := range contents {
		elementPath := path.Join(listPath, element.Name()) // f.e. list/element1 for folders, and list/element1.yaml for files
		if path.Ext(elementPath) == templateExtension {
			if debug {
				log.Print("Skipping object from '" + elementPath + "' since it is a template ...")
			}
		} else {
			if debug {
				log.Print("Loading object from '" + elementPath + "' ...")
			}
			if element.IsDir() {
				tempMappedObject := loadYaml(path.Join(elementPath, "index.yaml")) // f.e. list/element1/index.yaml
				tempMappedObject["Path"] = "/" + elementPath                       // f.e. /list/element1
				mappedObjects[elementPath] = tempMappedObject
			} else {
				tempMappedObject := loadYaml(elementPath) // f.e. list/element1.yaml
				tempMappedObject["Path"] = "/" + strings.TrimSuffix(elementPath, filepath.Ext(elementPath))
				// If the object is defined via a single file, no (further) path exists
				mappedObjects[elementPath] = tempMappedObject
			}
		}
	}

	return mappedObjects
}

func main() {
	// #####
	// START declaring variables
	// #####

	// no log.Println for debug before this, because the flags have to be read first ;)
	readCliFlags()
	// # example $> ./template -valuesfile values.yaml -inputDir ./ -partialsDir partials-html/ -templateExtension .html.template -generatedExtension .html

	if debug {
		log.Println("valuesFilePaths:", valuesFilePaths)
		log.Println("inputDir:", inputDir)
		log.Println("partialsDir:", partialsDir)
		log.Println("outputDir:", outputDir)
		log.Println("templateExtension:", templateExtension)
		log.Println("partialExtension:", partialExtension)
		log.Println("generatedExtension:", generatedExtension)
		log.Println("staticDir:", staticDir)
		log.Println("watch:", watch)
	}

	// #####
	// END declaring variables
	// START rendering
	// #####

	if !watch { // if not watching
		rebuildOutput() // delete old contents of output-folder & copy static contents & render templates once
	} else { // else (== if watching)
		watchAll() // start to watch
	}

	// #####
	// END rendering
	// #####
}
