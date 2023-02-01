package fileIO

import (
	"strings"
)

// Retains all filePaths that have the exact specified level from the root position.
// Works by counting slashes.
func (fileList FileList) FilterByLevel(level int) FileList {
	return fileList.Filter(
		func(s string) bool {
			return (strings.Count(s, "/") == level)
		})
}
