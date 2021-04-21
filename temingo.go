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
	"github.com/PuerkitoBio/purell"
	"github.com/imdario/mergo"
	"github.com/otiai10/copy"
	"github.com/rjeczalik/notify"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	debug bool
	watch bool

	valuesFilePaths         []string
	inputDir                string
	partialsDir             string
	outputDir               string
	staticDir               string
	templateExtension       string
	singleTemplateExtension string
	partialExtension        string
	generatedExtension      string

	listListObjects = make(map[string]map[string]interface{})
)

func createFolderIfNotExists(path string) {
	os.MkdirAll(path, os.ModePerm)
}

func createBreadcrumbs(path string) map[string]string {
	breadcrumbs := make(map[string]string)
	currentPath := ""
	dirNames := strings.Split(filepath.Dir(path), "/")
	for ok := true; ok; ok = (len(dirNames) != 0) {
		currentPath = currentPath + "/" + dirNames[0]
		breadcrumbs[dirNames[0]] = currentPath
		dirNames = dirNames[1:] // remove first one, as it is now added to 'currentPath'
	}

	return breadcrumbs
}

func isExcluded(path string, exclusions []string) bool {
	for _, exclusion := range exclusions {
		if strings.Contains(exclusion, "**") { // f.e. "**/file.a.b.c"
			splittedExclusion := strings.SplitAfter(exclusion, "**/")
			pattern := splittedExclusion[len(splittedExclusion)-1]
			basePath := filepath.Base(path)
			isMatch, err := filepath.Match(pattern, basePath)
			if err != nil {
				log.Fatal(err)
			}
			if isMatch {
				return true
			}
		} else { // f.e. "/*/*.file.a.b.c"
			isMatch, err := filepath.Match(exclusion, path)
			if err != nil {
				log.Fatal(err)
			}
			if isMatch {
				return true
			}
		}
	}

	return false
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
			if !isExcluded(entryPath, exclusions) {
				if entry.IsDir() {
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
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
		"list": func(listPaths ...string) map[string]interface{} {
			listObjects := make(map[string]interface{})
			if len(listPaths) == 0 { // If no path is provided
				listPaths = append(listPaths, filepath.Dir(name)) // Add the default path (folder containing the template)
			}
			for _, listPath := range listPaths {
				mergo.Merge(&listObjects, loadListObjects(listPath))
				listListObjects[listPath] = listObjects
			}
			return listObjects
		},
		"urlize": func(oldContent string) string {
			newContent, err := purell.NormalizeURLString(strings.ReplaceAll(oldContent, " ", "_"), purell.FlagsSafe)
			if err != nil {
				log.Fatal(err)
			}
			newContent = strings.ToLower(newContent) // Also convert everything to lowercase. Arguable.
			if debug {
				log.Println("Urlized '" + oldContent + "' to '" + newContent + "'.")
			}
			return newContent
		},
		"capitalize": func(oldContent string) string {
			newContent := strings.Title(oldContent)
			if debug {
				log.Println("Capitalized '" + oldContent + "' to '" + newContent + "'.")
			}
			return newContent
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
	createFolderIfNotExists(dirPath)
	err := ioutil.WriteFile(filePath, content, os.ModePerm)
	return err
}

func readCliFlags() {
	var (
		info fs.FileInfo
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
	flag.StringVarP(&generatedExtension, "generatedExtension", "g", "", "Sets the extension of the generated files.")
	flag.BoolVarP(&watch, "watch", "w", false, "Watches the template-file-directory, partials-directory and values-files.")
	flag.BoolVarP(&debug, "debug", "d", false, "Enables the debug mode.")

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

func getMappedValues() map[string]interface{} {
	var mappedValues map[string]interface{}
	for _, v := range valuesFilePaths {
		tempMappedValues := loadYaml(v)

		err := mergo.Merge(&mappedValues, tempMappedValues, mergo.WithOverride)
		if err != nil {
			log.Fatal(err)
		}
	}
	return mappedValues
}

func runTemplate(mappedValues map[string]interface{}, templateName string, template string, partialTemplates [][]string, outputFilePath string) {
	outputBuffer := new(bytes.Buffer)
	outputBuffer.Reset()
	tpl := parseTemplateFiles(templateName, template, partialTemplates)
	mappedValues["breadcrumbs"] = createBreadcrumbs(filepath.Dir(templateName))
	err := tpl.Execute(outputBuffer, mappedValues)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(outputDir); os.IsNotExist(err) { // If output directory doesn't exist
		createFolderIfNotExists(outputDir)
	}
	err = writeTemplateToFile(outputFilePath, outputBuffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}

func render() {
	// #####
	// START reading value files
	// #####
	if debug {
		log.Println("*** Reading values file(s) ... ***")
	}
	mappedValues := getMappedValues()
	if debug {
		valuesYaml, err := yaml.Marshal(mappedValues)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("*** General values-object: ***\n" + string(valuesYaml))
	}

	// #####
	// END reading value files
	// START normal templating
	// #####

	templates := getTemplates(inputDir, templateExtension, []string{
		path.Join(inputDir, partialsDir, "*"),
		path.Join(inputDir, outputDir, "*"),
		"**/*" + singleTemplateExtension,
	}) // get full html templates - with names
	partialTemplates := getTemplates(path.Join(inputDir, partialsDir), partialExtension, []string{path.Join(inputDir, outputDir)}) // get partial html templates - without names

	for _, template := range templates {
		outputFilePath := path.Join(outputDir, strings.TrimSuffix(template[0], templateExtension)+generatedExtension)
		if debug {
			log.Println("Writing output file '" + outputFilePath + "' ...")
		}
		runTemplate(mappedValues, template[0], template[1], partialTemplates, outputFilePath)
	}

	// #####
	// END normal templating
	// START single-view templating
	// #####

	// identify & collect single-view templates via their extension
	singleTemplates := getTemplates(inputDir, singleTemplateExtension, []string{
		path.Join(inputDir, partialsDir, "*"),
		path.Join(inputDir, outputDir, "*"),
	}) // get full html templates - with names

	// for each of the single-view templates
	for _, template := range singleTemplates {
		templateName := template[0]
		template := template[1]
		// search all configurations

		dirContents, err := ioutil.ReadDir(filepath.Dir(templateName))
		if err != nil {
			log.Fatal(err)
		}

		itemValues := make(map[string]interface{})

		// Read item-specific values, so they are available independent of the items way of the configuration
		for _, dirEntry := range dirContents {
			if !dirEntry.IsDir() && filepath.Ext(dirEntry.Name()) == ".yaml" {
				itemValues[path.Join(filepath.Dir(templateName), dirEntry.Name())] = loadYaml(path.Join(filepath.Dir(templateName), dirEntry.Name()))
			} else if dirEntry.IsDir() {
				if _, err := os.Stat(path.Join(filepath.Dir(templateName), dirEntry.Name(), "index.yaml")); err == nil { // if the dirEntry-folder contains an "index.yaml"
					itemValues[path.Join(filepath.Dir(templateName), dirEntry.Name())] = loadYaml(path.Join(filepath.Dir(templateName), dirEntry.Name(), "index.yaml"))
				}
			}
		}

		for itemPath, itemValue := range itemValues {
			// load corresponding additional values into mappedValues["Item"]
			extendedMappedValues := mappedValues
			itemPath = strings.TrimSuffix(itemPath, filepath.Ext(itemPath))
			fileName := strings.TrimSuffix(filepath.Base(templateName), singleTemplateExtension)
			extendedMappedValues["ItemPath"] = itemPath
			extendedMappedValues["Item"] = itemValue
			outputFilePath := path.Join(outputDir, itemPath, fileName)
			if debug {
				log.Println("Writing single-view output file '" + outputFilePath + "' ...")
			}
			runTemplate(extendedMappedValues, templateName, template, partialTemplates, outputFilePath)
		}
	}

	// #####
	// END single-view templating
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
			log.Println("Deleting output-dir content at: " + elementPath)
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
		log.Println("singleTemplateExtension:", singleTemplateExtension)
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
