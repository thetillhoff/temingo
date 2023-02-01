package fileIO

import (
	gitignore "github.com/sabhiram/go-gitignore"
)

// Removes all filePaths from fileList.Files that match the specified ignoreLines.
// IgnoreLines are in the format as they are in any .gitignore.
func (fileList FileList) FilterByIgnoreLines(ignoreLines []string) FileList {
	var ignore *gitignore.GitIgnore = gitignore.CompileIgnoreLines(ignoreLines...)

	return fileList.Filter(
		func(s string) bool {
			return !ignore.MatchesPath(s)
		})
}
