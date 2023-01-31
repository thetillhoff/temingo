package fileIO

// Implicitly tested by all the other included Filters
func (fileList FileList) Filter(filterFunc func(string) bool) FileList {
	var (
		include bool
		files   []string = []string{}
	)

	for _, filePath := range fileList.Files {
		include = filterFunc(filePath)
		if include {
			files = append(files, filePath)
		}
	}

	fileList.Files = files
	return fileList
}
