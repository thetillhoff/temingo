package fileIO

import (
	"strings"
)

func (fileList *FileList) RemovePrefixFromAllPaths(prefix string) *FileList {
	for _, filePath := range fileList.Files {
		filePath = strings.TrimPrefix(filePath, prefix)
	}

	return fileList
}
