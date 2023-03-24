package common

import (
	"path/filepath"
	"strings"
)

func ValidateFilePath(path string) string {
	path = filepath.Clean(path)
	path = filepath.ToSlash(path)
	path = strings.TrimLeft(path, "/")

	return path
}

func DirOfFilePath(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		dir = ""
	}

	return dir
}
