package fileIO

import (
	"strings"
)

// Retains the filepaths below the provided path, which are the specified amount of path-levels below it
// Requires a trailing slash. If there is none, it will be automatically added temporarily.
func (fileList FileList) FilterByLevelAtFolderPath(path string, level int) FileList {
	path = ensureTrailingSlash(path)
	return fileList.FilterByFolderPath(path).Filter(
		func(s string) bool {
			trimmed := strings.TrimPrefix(s, path)
			return (strings.Count(trimmed, "/") == level)
		})
}
