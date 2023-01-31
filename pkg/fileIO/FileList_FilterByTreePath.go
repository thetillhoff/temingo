package fileIO

import (
	"path"
	"strings"
)

func (fileList FileList) FilterByTreePath(treepath string) FileList {
	var (
		files         []string
		folders       []string
		currentFolder string = ""
	)

	treepath = ensureNoTrailingSlash(treepath) // If there would be a trailing slash, it would add the last element twice

	files = fileList.FilterByLevel(0).Files // Add contents of root dir aka ""

	folders = strings.Split(treepath, "/")

	for _, folder := range folders { // For each subfolder in the treepath
		currentFolder = path.Join(currentFolder, folder)
		files = append(files, fileList.FilterByLevelAtFolderPath(currentFolder, 0).Files...) // Add all files in that subfolder to fileList
	}

	fileList.Files = files

	return fileList
}
