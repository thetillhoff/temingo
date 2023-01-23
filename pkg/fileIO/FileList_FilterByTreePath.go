package fileIO

import (
	"path"
	"strings"
)

func (fileList *FileList) FilterByTreePath(treepath string) *FileList {
	var (
		files         []string
		folders       []string
		currentFolder string = ""
	)

	files = fileList.FilterByLevel(0).Files // root dir aka ""

	folders = strings.Split(treepath, "/")

	for _, folder := range folders {
		currentFolder = path.Join(currentFolder, folder)
		files = append(fileList.FilterByLevelAtPath(currentFolder, 0).Files)
	}

	fileList.Files = files

	return fileList
}
