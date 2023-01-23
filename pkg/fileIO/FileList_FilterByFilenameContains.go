package fileIO

import (
	"path"
	"strings"
)

func (fileList *FileList) FilterByFileNameContains(nameSubstring string) *FileList {
	return fileList.Filter(
		func(s string) bool {
			return (strings.Contains(path.Base(s), nameSubstring))
		})
}
