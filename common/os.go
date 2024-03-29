package common

import (
	"os"
	"syscall"
	"time"
)

const (
	NUL   byte = 0x00
	SPACE byte = 0x20
	DELIM byte = 0x0a
)

func StatTimes(fi os.FileInfo) (atime, mtime, ctime time.Time) {
	mtime = fi.ModTime()
	stat := fi.Sys().(*syscall.Stat_t)
	atime = time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec))
	ctime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	return
}
