package fileIO

import (
	"path"
)

func (fileList FileList) FilterByFilename(name string) FileList {
	return fileList.Filter(
		func(s string) bool {
			return (path.Base(s) == name)
		})
}
