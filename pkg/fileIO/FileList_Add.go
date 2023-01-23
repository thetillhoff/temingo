package fileIO

func (fileList *FileList) Add(filePathsToAdd ...string) *FileList {

	fileList.Files = append(fileList.Files, filePathsToAdd...)

	return fileList
}
