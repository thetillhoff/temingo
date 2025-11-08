package temingo

import (
	"log/slog"
	"os"
)

// Returns the available ProjectTypes for the Init() function
func ProjectTypes() []string { // "static" slice
	var typeList []string

	dirEntries, err := embeddedExampleProjectFilesWithPrefix.ReadDir("InitFiles")
	if err != nil {
		slog.Default().Error("Failed to read InitFiles directory", "error", err)
		os.Exit(1)
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			typeList = append(typeList, dirEntry.Name())
		}
	}

	return typeList
}
