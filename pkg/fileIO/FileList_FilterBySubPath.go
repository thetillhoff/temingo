package fileIO

import (
	"strings"
)

// Returns all paths in FileList that match a specific subpath.
// If the path doesn't exist, the result is empty.
func (fileList *FileList) FilterBySubPath(subPath string) *FileList {
	return fileList.Filter(
		func(s string) bool {
			return strings.HasPrefix(s, subPath)
		})
}
