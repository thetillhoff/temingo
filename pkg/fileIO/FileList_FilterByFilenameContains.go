package fileIO

import (
	"path"
	"strings"
)

// Retain only those filePaths, where the fileName contains the specified subString
func (fileList FileList) FilterByFilenameContains(nameSubstring string) FileList {
	return fileList.Filter(
		func(s string) bool {
			return (strings.Contains(path.Base(s), nameSubstring))
		})
}
