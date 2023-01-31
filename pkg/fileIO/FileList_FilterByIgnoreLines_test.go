package fileIO

import (
	"strconv"
	"testing"
)

// Test if only the paths are returned that are not matched against the ignore file lines
func TestFilterByIgnoreLines(t *testing.T) {
	expectedValues := []string{"random.file", "folder/with/subfolder/random.file"}

	fileList := FileList{Files: []string{
		"random.file",
		"ignored.file",
		"folder/with/subfolder/random.file",
		"folder/with/subfolder/ignored.file",
		"ignoredfolder/random.file",
		"somwhere/ignoredfolder/random.file",
	}}

	fileList = fileList.FilterByIgnoreLines([]string{"ignored.file", "ignoredfolder/**"})

	if len(fileList.Files) != len(expectedValues) {
		t.Fatal("expected length of fileList.Files is", len(expectedValues), "got", len(fileList.Files), ":\n", fileList.Files)
	}

	for index, actualValue := range fileList.Files {
		if expectedValues[index] != actualValue {
			t.Fatal("expected value of fileList.Files["+strconv.Itoa(index)+"] is", expectedValues[index], "got", actualValue)
		}
	}
}
