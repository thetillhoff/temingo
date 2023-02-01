package fileIO

// Adds the provided filePaths to the fileList.Files
func (fileList FileList) Add(filePathsToAdd ...string) FileList {

	fileList.Files = append(fileList.Files, filePathsToAdd...)

	return fileList
}
