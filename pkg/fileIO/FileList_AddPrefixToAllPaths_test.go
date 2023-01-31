package fileIO

import (
	"strconv"
	"testing"
)

// Test if the Element is correctly added when the FileList is previously not empty
func TestAddPrefixToAllPaths(t *testing.T) {
	expectedValues := []string{"prefixed/folder/some.file", "prefixed/folder/subfolder/with/another.file"}

	fileList := FileList{Files: []string{
		"some.file",
		"subfolder/with/another.file",
	}}

	fileList = fileList.AddPrefixToAllPaths("prefixed/folder")

	if len(fileList.Files) != len(expectedValues) {
		t.Fatal("expected length of fileList.Files is", len(expectedValues), "got", len(fileList.Files), ":\n", fileList.Files)
	}

	for index, actualValue := range fileList.Files {
		if expectedValues[index] != actualValue {
			t.Fatal("expected value of fileList.Files["+strconv.Itoa(index)+"] is", expectedValues[index], "got", actualValue)
		}
	}
}
