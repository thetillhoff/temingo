package fileIO

// Removes the provided filePaths to the fileList.Files
func (fileList FileList) Remove(filePathsToRemove ...string) FileList {

	var filePaths = []string{}

	for _, filePath := range fileList.Files {
		remove := false
		for _, filePathToRemove := range filePathsToRemove {
			if filePath == filePathToRemove {
				remove = true
				break
			}
		}
		if !remove {
			filePaths = append(filePaths, filePath)
		}
	}

	fileList.Files = filePaths

	return fileList
}
