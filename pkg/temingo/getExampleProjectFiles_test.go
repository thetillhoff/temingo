package temingo

import (
	"testing"
)

// Check if returned project files contains at least one of the expected files if the passed type is valid
func TestGetExampleProjectFilesWithValidProjectType(t *testing.T) {
	expectedPath := "src/index.template.html"

	engine := DefaultEngine()

	exampleProjectFiles, err := engine.getExampleProjectFiles("example")
	if err != nil {
		t.Fatal("expected example project file retrieval for existing project type to succeed, got error:", err)
	}

	contains := false
	for path := range exampleProjectFiles {
		if path == expectedPath {
			contains = true
			break
		}
	}
	if !contains {
		t.Fatal("expected example project file to contain file at", expectedPath, "got", exampleProjectFiles)
	}

}

// Check if the function fails if the passed type is invalid
func TestGetExampleProjectFilesWithInValidProjectType(t *testing.T) {

	engine := DefaultEngine()

	_, err := engine.getExampleProjectFiles("invalid")

	if err == nil {
		t.Fatal("expected example project file retrieval for inexisting project type to fail, got success")
	}

}
