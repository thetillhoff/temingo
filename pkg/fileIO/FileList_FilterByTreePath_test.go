package fileIO

import (
	"strconv"
	"testing"
)

// Test if only the paths are returned that match a specified treepath
func TestFilterByTreePath(t *testing.T) {
	expectedValues := []string{"folderA/random.file", "folderA/subfolderA/random.file"}

	fileList := FileList{Files: []string{
		"random.file",
		"folderA/random.file",
		"folderA/subfolderA/random.file",
		"folderB/random.file",
		"folderB/subfolder/random.file",
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
