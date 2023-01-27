package utils

import "os"

/*
FileExists checks if a regular file (not directory) with a given filepath exist
*/
func FileExists(filepath string) bool {

	fileinfo, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}
	// Return false if the fileinfo says the file path is a directory.
	return !fileinfo.IsDir()
}

func DirectoryExists(filepath string) bool {

	fileinfo, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}
	// Return true if the fileinfo says the file path is a directory.
	return fileinfo.IsDir()
}
