package common

import (
	"fmt"
	"io/fs"
	"strconv"
)

type FileMode = fs.FileMode
type FileInfo = fs.FileInfo

const (
	Empty FileMode = 0

	Dir FileMode = 0040000

	Regular FileMode = 0100644

	Deprecated FileMode = 0100664

	Executable FileMode = 0100755

	Symlink FileMode = 0120000

	Submodule FileMode = 0160000
)

// New takes the octal string representation of a FileMode and returns
// the FileMode and a nil error.  If the string can not be parsed to a
// 32 bit unsigned octal number, it returns Empty and the parsing error.
//
// Example: "40000" means Dir, "100644" means Regular.
func NewFileMode(s string) (FileMode, error) {
	n, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return Empty, err
	}

	return FileMode(n), nil
}

func FileModeToString(m FileMode) string {
	return fmt.Sprintf("%06o", uint32(m))
}

func IsRegular(m FileMode) bool {
	return m == Regular ||
		m == Deprecated
}

func IsFile(m FileMode) bool {
	return m == Regular ||
		m == Deprecated ||
		m == Executable ||
		m == Symlink
}

func IsDir(m FileMode) bool {
	return m == Dir
}
