package temingo

import (
	"strconv"
	"testing"
)

// Check if ProjectTypes list is of expected length
func TestProjectTypesLength(t *testing.T) {
	expectedValue := 2

	projectTypes := ProjectTypes()
	if len(projectTypes) != expectedValue {
		t.Fatal("wrong amount of projectTypes returned from ProjectTypes: expected"+strconv.Itoa(expectedValue)+", got", len(projectTypes), ", contents:", projectTypes) // printing contents as well, so it's easier to see where the error might come from
	}
}

// Check if ProjectTypes list contains at least one of the expected types
func TestProjectTypesContent(t *testing.T) {
	expectedValue := "example"

	projectTypes := ProjectTypes()

	contains := false
	for _, projectType := range projectTypes {
		if projectType == expectedValue {
			contains = true
			break
		}
	}
	if !contains {
		t.Fatal("missing projectType: expected at least '"+expectedValue+"', got", projectTypes)
	}
}
