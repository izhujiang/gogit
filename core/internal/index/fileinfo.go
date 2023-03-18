package index

import (
	"time"

	"github.com/izhujiang/gogit/common"
)

type fileinfo struct {
	name  string
	cTime time.Time
	mTime time.Time

	// 32-bit dev (divice)
	dev uint32
	// 32-bit ino (inode)
	ino  uint32
	mode common.FileMode
	uid  uint32
	gid  uint32

	// File size on-disk size from stat(2), truncated to 32-bit.
	size uint32
}

type Stat_t struct {
	Dev        uint32
	Mode       uint32
	Ino        uint32
	Uid        uint32
	Ctime_Sec  uint32
	Ctime_Nsec uint32
	Mtime_Sec  uint32
	Mtime_Nsec uint32
	Gid        uint32
	Size       uint32
}

func (f *fileinfo) Name() string {
	return f.name
}

func (f *fileinfo) Size() int64 {
	return int64(f.size)
}

func (f *fileinfo) Mode() common.FileMode {
	return common.FileMode(f.mode)
}

func (f *fileinfo) ModTime() time.Time {
	return f.mTime
}

func (f *fileinfo) IsDir() bool {
	return common.IsDir(common.FileMode(f.mode))
}

// return a pointer to syscall.Stat_t struct
func (f *fileinfo) Sys() any {
	stat_t := &Stat_t{
		Dev:        f.dev,
		Mode:       uint32(f.mode),
		Ino:        f.ino,
		Uid:        f.uid,
		Ctime_Sec:  uint32(f.cTime.Second()),
		Ctime_Nsec: uint32(f.cTime.Nanosecond()),
		Mtime_Sec:  uint32(f.mTime.Second()),
		Mtime_Nsec: uint32(f.mTime.Nanosecond()),
		Gid:        f.gid,
		Size:       f.size,
	}

	return stat_t
}
