package fileIO

import (
	"strconv"
	"testing"
)

// Test if the Element is correctly added when the FileList is previously empty
func TestAddWithEmpty(t *testing.T) {
	expectedLength := 1
	expectedValue := "test/foo.bar"

	fileList := FileList{Files: []string{}}
	fileList = fileList.Add("test/foo.bar") // TODO would be nice to be able to do just fileList = fileList.Add()...

	if len(fileList.Files) != expectedLength {
		t.Fatal("expected length of fileList.Files is", expectedLength, "got", len(fileList.Files))
	}

	if fileList.Files[0] != "test/foo.bar" {
		t.Fatal("expected value of fileList.Files[0] is", expectedValue, "got", fileList.Files[0])
	}
}

// Test if the Element is correctly added when the FileList is previously not empty
func TestAddWithNonempty(t *testing.T) {
	expectedValues := []string{"pre/contained.file", "test/foo.bar"}

	fileList := FileList{Files: []string{
		"pre/contained.file",
	}}

	fileList = fileList.Add("test/foo.bar")

	if len(fileList.Files) != len(expectedValues) {
		t.Fatal("expected length of fileList.Files is", len(expectedValues), "got", len(fileList.Files))
	}

	for index, actualValue := range fileList.Files {
		if expectedValues[index] != actualValue {
			t.Fatal("expected value of fileList.Files["+strconv.Itoa(index)+"] is", expectedValues[index], "got", actualValue)
		}
	}
}
