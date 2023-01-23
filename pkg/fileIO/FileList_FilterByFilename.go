package fileIO

import (
	"path"
)

func (fileList *FileList) FilterByFileName(name string) *FileList {
	return fileList.Filter(
		func(s string) bool {
			return (path.Base(s) == name)
		})
}
