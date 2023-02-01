package temingo

import "log"

// Returns the available ProjectTypes for the Init() function
func ProjectTypes() []string { // "static" slice
	var typeList []string

	dirEntries, err := embeddedExampleProjectFilesWithPrefix.ReadDir("InitFiles")
	if err != nil {
		log.Fatalln(err)
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			typeList = append(typeList, dirEntry.Name())
		}
	}

	return typeList
}
