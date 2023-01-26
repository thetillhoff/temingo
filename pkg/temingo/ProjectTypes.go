package temingo

import "log"

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
