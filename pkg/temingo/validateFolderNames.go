package temingo

import (
	"fmt"
	"path"
	"strings"

	"github.com/thetillhoff/fileIO"
)

// invalidFolderChars contains characters that break URLs when unencoded in a path segment.
const invalidFolderChars = " #?&<>\"{}|\\^`%"

// validateFolderNames checks all directory components in the file list for URL-breaking characters.
func validateFolderNames(fileList fileIO.FileList) error {
	seen := map[string]bool{}
	for _, filePath := range fileList.Files {
		dir := path.Dir(filePath)
		if dir == "." {
			continue
		}
		for _, segment := range strings.Split(dir, "/") {
			if segment == "" || segment == "." || seen[segment] {
				continue
			}
			seen[segment] = true
			if idx := strings.IndexAny(segment, invalidFolderChars); idx != -1 {
				return fmt.Errorf("invalid character %q in folder name %q (path: %s)", string(segment[idx]), segment, filePath)
			}
		}
	}
	return nil
}
