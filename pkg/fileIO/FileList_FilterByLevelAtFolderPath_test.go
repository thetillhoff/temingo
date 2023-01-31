package fileIO

import (
	"strconv"
	"testing"
)

// Test if only the paths are returned that match a specific level, starting from a specific path
func TestFilterByLevelAtFolderPathWithTrailingSlash(t *testing.T) {
	expectedValues := []string{"folder/subfolder/random.file"}

	fileList := FileList{Files: []string{
		"random.file",
		"folder/random.file",
		"folder/subfolder/random.file",
		"folder/subfolder/subsubfolder/random.file",
	}}

	fileList = fileList.FilterByLevelAtFolderPath("folder/", 1)

	if len(fileList.Files) != len(expectedValues) {
		t.Fatal("expected length of fileList.Files is", len(expectedValues), "got", len(fileList.Files), ":\n", fileList.Files)
	}

	for index, actualValue := range fileList.Files {
		if expectedValues[index] != actualValue {
			t.Fatal("expected value of fileList.Files["+strconv.Itoa(index)+"] is", expectedValues[index], "got", actualValue)
		}
	}
}
