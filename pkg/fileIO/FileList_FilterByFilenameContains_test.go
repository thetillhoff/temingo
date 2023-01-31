package fileIO

import (
	"strconv"
	"testing"
)

// Test if only the paths are returned that point to a matching filename
// Also test if matches are only verified against filenames and not the whole path
func TestFilterByFilenameContains(t *testing.T) {
	expectedValues := []string{"prefixed-matching.file", "folder/with/subfolder/matching.file.suffix"}

	fileList := FileList{Files: []string{
		"random.file",
		"prefixed-matching.file",
		"folder/with/subfolder/matching.file.suffix",
		"folder/with/subfolder/random.file",
		"matching.file/that/is/actually/a.folder",
	}}

	fileList = fileList.FilterByFilenameContains("matching.file")

	if len(fileList.Files) != len(expectedValues) {
		t.Fatal("expected length of fileList.Files is", len(expectedValues), "got", len(fileList.Files), ":\n", fileList.Files)
	}

	for index, actualValue := range fileList.Files {
		if expectedValues[index] != actualValue {
			t.Fatal("expected value of fileList.Files["+strconv.Itoa(index)+"] is", expectedValues[index], "got", actualValue)
		}
	}
}
