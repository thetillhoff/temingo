package fileIO

import (
	"strings"
)

// Removes the specified prefix to every filePath in fileList.Files
func (fileList FileList) RemovePrefixFromAllPaths(prefix string) FileList {
	for index := range fileList.Files {
		fileList.Files[index] = strings.TrimPrefix(fileList.Files[index], prefix)
	}

	return fileList
}
