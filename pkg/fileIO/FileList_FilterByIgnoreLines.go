package fileIO

import (
	gitignore "github.com/sabhiram/go-gitignore"
)

func (fileList *FileList) FilterByIgnoreLines(ignoreLines []string) *FileList {
	var (
		ignore *gitignore.GitIgnore
	)
	ignore = gitignore.CompileIgnoreLines(ignoreLines...)

	return fileList.Filter(
		func(s string) bool {
			return !ignore.MatchesPath(s)
		})
}
