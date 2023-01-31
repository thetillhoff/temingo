package fileIO

import (
	"strconv"
	"testing"
)

// Test if only the paths are returned that match a specified treepath
func TestFilterByTreePath(t *testing.T) {
	expectedValues := []string{"match.file", "folderA/match.file", "folderA/subfolderA/match.file"}

	fileList := FileList{Files: []string{
		"match.file",
		"folderA/match.file",
		"folderA/subfolderA/match.file",
		"folderB/random.file",
		"folderB/subfolderA/random.file",
		"folderA/subfolderB/random.file",
		"folderA/subfolderA/onemore/random.file",
	}}

	fileList = fileList.FilterByTreePath("folderA/subfolderA")

	if len(fileList.Files) != len(expectedValues) {
		t.Fatal("expected length of fileList.Files is", len(expectedValues), "got", len(fileList.Files), ":\n", fileList.Files)
	}

	for index, actualValue := range fileList.Files {
		if expectedValues[index] != actualValue {
			t.Fatal("expected value of fileList.Files["+strconv.Itoa(index)+"] is", expectedValues[index], "got", actualValue)
		}
	}
}
