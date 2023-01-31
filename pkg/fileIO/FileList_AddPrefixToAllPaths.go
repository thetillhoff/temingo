package fileIO

import "path"

func (fileList FileList) AddPrefixToAllPaths(prefix string) FileList {
	for index := range fileList.Files {
		fileList.Files[index] = path.Join(prefix, fileList.Files[index])
	}

	return fileList
}
