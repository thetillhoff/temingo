package temingo

import (
	"testing"
)

// Check if component verification succeeds if all component names are unique
func TestVerifyComponentsAllUnique(t *testing.T) {

	componentFiles := map[string]string{} // map(path)content

	componentFiles["src/components/a.component.html"] = `
	{{ define "testComponent1" }}
	test
	{{ end }}
	`
	componentFiles["src/components/b.component.html"] = `
	{{ define "testComponent2" }}
	test
	{{ end }}

	{{ define "testComponent3" }}
	test
	{{ end }}
	`

	err := verifyComponents(componentFiles) // Check if the components are unique

	if err != nil {
		t.Fatal("expected component verification to succeed, got error:", err)
	}

}

// Check if component verification fails if not all component names in one of the files are unique
func TestVerifyComponentsNonUniqueOneFile(t *testing.T) {

	componentFiles := map[string]string{} // map(path)content

	componentFiles["src/components/a.component.html"] = `
	{{ define "testComponent1" }}
	test
	{{ end }}
	`
	componentFiles["src/components/b.component.html"] = `
	{{ define "testComponent2" }}
	test
	{{ end }}

	{{ define "testComponent2" }}
	test
	{{ end }}
	`

	err := verifyComponents(componentFiles) // Check if the components are unique

	if err == nil {
		t.Fatal("expected component verification to fail, got success")
	}

}

// Check if component verification fails if not all component names in multiple files are unique
func TestVerifyComponentsNonUniqueMultipleFiles(t *testing.T) {

	componentFiles := map[string]string{} // map(path)content

	componentFiles["src/components/a.component.html"] = `
	{{ define "testComponent1" }}
	test
	{{ end }}
	`
	componentFiles["src/components/b.component.html"] = `
	{{ define "testComponent2" }}
	test
	{{ end }}

	{{ define "testComponent1" }}
	test
	{{ end }}
	`

	err := verifyComponents(componentFiles) // Check if the components are unique

	if err == nil {
		t.Fatal("expected component verification to fail, got success")
	}

}
