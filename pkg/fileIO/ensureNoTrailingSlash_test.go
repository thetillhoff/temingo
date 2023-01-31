package fileIO

import (
	"testing"
)

// Test if the Slash is not removed if it's not there
func TestNoTrailingSlashWithNoSlash(t *testing.T) {
	expectedValue := "foo/bar"

	result := ensureNoTrailingSlash("foo/bar")

	if result != expectedValue {
		t.Fatal("expected value of ensureTrailingSlash is", expectedValue, "got", result)
	}
}

// Test if the Slash is removed if it's there
func TestNoTrailingSlashWithExistingSlash(t *testing.T) {
	expectedValue := "foo/bar"

	result := ensureNoTrailingSlash("foo/bar/")

	if result != expectedValue {
		t.Fatal("expected value of ensureTrailingSlash is", expectedValue, "got", result)
	}
}
