package fileIO

import (
	"strconv"
	"testing"
)

// Test if only the paths are returned that point to a matching filename
func TestFilterByFilename(t *testing.T) {
	expectedValues := []string{"matching.file", "folder/with/subfolder/matching.file"}

	fileList := FileList{Files: []string{
		"not-matching.file",
		"matching.file",
		"folder/with/subfolder/matching.file",
		"folder/with/subfolder/not-matching.file",
	}}

	fileList = fileList.FilterByFilename("matching.file")

	if len(fileList.Files) != len(expectedValues) {
		t.Fatal("expected length of fileList.Files is", len(expectedValues), "got", len(fileList.Files), ":\n", fileList.Files)
	}

	for index, actualValue := range fileList.Files {
		if expectedValues[index] != actualValue {
			t.Fatal("expected value of fileList.Files["+strconv.Itoa(index)+"] is", expectedValues[index], "got", actualValue)
		}
	}
}
