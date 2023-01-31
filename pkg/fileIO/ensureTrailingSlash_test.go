package fileIO

import (
	"testing"
)

// Test if the Slash is added if it's not there yet
func TestTrailingSlashWithNoSlash(t *testing.T) {
	expectedValue := "foo/bar/"

	result := ensureTrailingSlash("foo/bar")

	if result != expectedValue {
		t.Fatal("expected value of ensureTrailingSlash is", expectedValue, "got", result)
	}
}

// Test if the Slash is not added if it's already there yet
func TestTrailingSlashWithExistingSlash(t *testing.T) {
	expectedValue := "foo/bar/"

	result := ensureTrailingSlash("foo/bar/")

	if result != expectedValue {
		t.Fatal("expected value of ensureTrailingSlash is", expectedValue, "got", result)
	}
}
