package temingo

import (
	"testing"
)

// Check if partial verification succeeds if all partial names are unique
func TestVerifyPartialsAllUnique(t *testing.T) {

	partialFiles := map[string]string{} // map(path)content

	partialFiles["src/partials/a.partial.html"] = `
	{{ define "testPartial1" }}
	test
	{{ end }}
	`
	partialFiles["src/partials/b.partial.html"] = `
	{{ define "testPartial2" }}
	test
	{{ end }}

	{{ define "testPartial3" }}
	test
	{{ end }}
	`

	err := verifyPartials(partialFiles) // Check if the partials are unique

	if err != nil {
		t.Fatal("expected partial verification to succeed, got error:", err)
	}

}

// Check if partial verification fails if not all partial names in one of the files are unique
func TestVerifyPartialsNonUniqueOneFile(t *testing.T) {

	partialFiles := map[string]string{} // map(path)content

	partialFiles["src/partials/a.partial.html"] = `
	{{ define "testPartial1" }}
	test
	{{ end }}
	`
	partialFiles["src/partials/b.partial.html"] = `
	{{ define "testPartial2" }}
	test
	{{ end }}

	{{ define "testPartial2" }}
	test
	{{ end }}
	`

	err := verifyPartials(partialFiles) // Check if the partials are unique

	if err == nil {
		t.Fatal("expected partial verification to fail, got success")
	}

}

// Check if partial verification fails if not all partial names in multiple files are unique
func TestVerifyPartialsNonUniqueMultipleFiles(t *testing.T) {

	partialFiles := map[string]string{} // map(path)content

	partialFiles["src/partials/a.partial.html"] = `
	{{ define "testPartial1" }}
	test
	{{ end }}
	`
	partialFiles["src/partials/b.partial.html"] = `
	{{ define "testPartial2" }}
	test
	{{ end }}

	{{ define "testPartial1" }}
	test
	{{ end }}
	`

	err := verifyPartials(partialFiles) // Check if the partials are unique

	if err == nil {
		t.Fatal("expected partial verification to fail, got success")
	}

}
