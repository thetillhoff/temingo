package fileIO

import (
	"strings"
)

func (fileList FileList) FilterByLevel(level int) FileList {
	return fileList.Filter(
		func(s string) bool {
			// TODO doesn't work fully yet, as it cannot filter for a level at a specific path
			return (strings.Count(s, "/") == level)
		})
}
