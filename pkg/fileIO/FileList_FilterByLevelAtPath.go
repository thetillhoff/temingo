package fileIO

import (
	"strings"
)

func (fileList *FileList) FilterByLevelAtPath(path string, level int) *FileList {
	return fileList.FilterBySubPath(path).Filter(
		func(s string) bool {
			trimmed := strings.TrimPrefix(s, path)
			return (strings.Count(trimmed, "/") == level)
		})
}