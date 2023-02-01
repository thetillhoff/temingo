package fileIO

// Filters the fileList based on the provided function.
// The provided function gets each filePath as argument and should return whether it should be included or not.
// Inclusion means it is still contained in the returned fileList
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
