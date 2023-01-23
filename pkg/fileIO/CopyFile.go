package fileIO

import (
	"io"
	"os"
	"path"
)

// Copies the specified sourcePath to the destinationPath
// Will NOT create the required folders.
// If the specified sourcePath is a folder, it will create the folder at the destination, but not copy the contents (== non-recursive).
// This copy process respects permissions of the sourcePaths while creating the new files/folders.
// The destinationPath must not exist.
func CopyFile(sourcePath string, destinationPath string) error {
	var (
		err error
		// Read different recommendations about the buffer size;
		// 128KB==128*1024b: https://eklitzke.org/efficient-file-copying-on-linux
		// 4KB=4*1024b:      https://stackoverflow.com/questions/3033771/file-i-o-with-streams-best-memory-buffer-size
		buffer          = make([]byte, 4*1024)
		sourceFile      *os.File
		destinationFile *os.File
	)

	// Get information like permission about the sourcefile
	sourceFileStat, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if sourceFileStat.IsDir() { // Copying a folder
		err = os.Mkdir(destinationPath, sourceFileStat.Mode().Perm()) // Creating destination folder with same permissions as source folder
		if err != nil {
			return err
		}
	} else { // Copying a file
		sourceFile, err = os.Open(sourcePath) // Open sourcefile, so it can be read from
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		err = os.MkdirAll(path.Dir(destinationPath), os.ModePerm) // Create containing folder for destination
		if err != nil {
			return err
		}

		destinationFile, err = os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, sourceFileStat.Mode().Perm()) // writeonly, create if not exists, not append (==overwrite), file must not exist
		if err != nil {
			return err
		}
		defer destinationFile.Close()

		for {
			n, err := sourceFile.Read(buffer)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 { // == err == io.EOF
				break
			}

			if _, err := destinationFile.Write(buffer[:n]); err != nil {
				return err
			}
		}
	}

	return nil
}
