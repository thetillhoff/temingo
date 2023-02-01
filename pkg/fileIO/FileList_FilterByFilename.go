package fileIO

import (
	"path"
)

// Retain those filePaths that have the specified fileName
func (fileList FileList) FilterByFilename(name string) FileList {
	return fileList.Filter(
		func(s string) bool {
			return (path.Base(s) == name)
		})
}
