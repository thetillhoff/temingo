package fileIO

import (
	"strings"
)

// Retains all paths in FileList that match a specific subpath.
// If the path doesn't exist, the result is empty.
// Requires a trailing slash. If there is none, it will be automatically added temporarily.
func (fileList FileList) FilterByFolderPath(subPath string) FileList {
	subPath = ensureTrailingSlash(subPath)
	return fileList.Filter(
		func(s string) bool {
			return strings.HasPrefix(s, subPath)
		})
}
