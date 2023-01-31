package fileIO

import (
	"strings"
)

func (fileList FileList) RemovePrefixFromAllPaths(prefix string) FileList {
	for index := range fileList.Files {
		fileList.Files[index] = strings.TrimPrefix(fileList.Files[index], prefix)
	}

	return fileList
}
