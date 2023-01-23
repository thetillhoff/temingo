package fileIO

func (fileList *FileList) AddPrefixToAllPaths(prefix string) *FileList {
	for _, filePath := range fileList.Files {
		filePath = filePath + prefix
	}

	return fileList
}
