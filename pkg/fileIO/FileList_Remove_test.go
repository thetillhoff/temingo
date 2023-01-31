package fileIO

import (
	"testing"
)

// Test if the element is correctly removed if it exists in a non-empty FileList
func TestRemoveWithMatchInNonEmpty(t *testing.T) {
	expectedLength := 1
	expectedValue := "unmatching.file"
	fileList := FileList{Files: []string{
		"matching.file",
		expectedValue,
	}}

	fileList.Files = fileList.Remove("matching.file").Files // TODO would be nice to be able to do just fileList = fileList.Remove()...

	if len(fileList.Files) != expectedLength {
		t.Fatal("expected length of fileList.Files is", expectedLength, "got", len(fileList.Files))
	}

	if fileList.Files[0] != expectedValue {
		t.Fatal("expected fileList.Files[0] to be", expectedValue, "got", fileList.Files[0])
	}
}

// Test if the list is still empty when it was empty before
func TestRemoveWithEmpty(t *testing.T) {
	expectedLength := 0

	fileList := FileList{Files: []string{}}
	fileList = fileList.Remove("test/foo.bar")

	if len(fileList.Files) != expectedLength {
		t.Fatal("expected length of fileList.Files is", expectedLength, "got", len(fileList.Files))
	}
}

// Test if the list is still the same when it doesn't contain the path to be removed.
func TestRemoveWithNonempty(t *testing.T) {
	expectedLength := 2
	fileList := FileList{Files: []string{
		"unmatching.file1",
		"unmatching.file2",
	}}

	fileList.Files = fileList.Remove("matching.file").Files // TODO would be nice to be able to do just fileList = fileList.Remove()...

	if len(fileList.Files) != expectedLength {
		t.Fatal("expected length of fileList.Files is", expectedLength, "got", len(fileList.Files), ":\n", fileList.Files)
	}
}
