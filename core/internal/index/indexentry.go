package index

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/izhujiang/gogit/common"
)

type IndexEntry struct {
	Oid      common.Hash
	Filepath string

	CTime time.Time
	MTime time.Time

	// 32-bit dev (divice)
	Dev uint32
	// 32-bit ino (inode)
	Ino   uint32
	Mode  common.FileMode
	Stage Stage
	Uid   uint32
	Gid   uint32

	// File size on-disk size from stat(2), truncated to 32-bit.
	Size uint32

	Skipworktree bool
	IntentToAdd  bool
}

func NewIndexEntry(oid common.Hash, mode common.FileMode, filepath string) *IndexEntry {
	ie := &IndexEntry{
		Oid:      oid,
		Mode:     mode,
		Filepath: filepath,
	}
	return ie
}

func NewIndexEntryWithFileInfo(oid common.Hash, mode common.FileMode, filepath string, fi os.FileInfo) *IndexEntry {
	e := &IndexEntry{
		Oid:      oid,
		Mode:     mode,
		Filepath: filepath,
	}

	stat := fi.Sys().(*syscall.Stat_t)
	e.Mode = common.FileMode(stat.Mode)
	e.CTime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	e.MTime = fi.ModTime()
	e.Dev = uint32(stat.Dev)
	e.Ino = uint32(stat.Ino)
	e.Uid = stat.Gid
	e.Gid = stat.Gid
	e.Size = uint32(stat.Size)
	return e
}
func (e *IndexEntry) UpdateWithFileInfo(fi os.FileInfo) {
	stat := fi.Sys().(*syscall.Stat_t)
	e.Mode = common.FileMode(stat.Mode)
	e.CTime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	e.MTime = fi.ModTime()
	e.Dev = uint32(stat.Dev)
	e.Ino = uint32(stat.Ino)
	e.Uid = stat.Gid
	e.Gid = stat.Gid
	e.Size = uint32(stat.Size)
}

type IndexEntries struct {
	entries []*IndexEntry
}

func (ide *IndexEntries) size() int {
	return len(ide.entries)
}
func (ide *IndexEntries) reset() {
	ide.entries = make([]*IndexEntry, 0, 256)
}

func (ide *IndexEntries) find(path string) *IndexEntry {
	for _, entry := range ide.entries {
		if entry.Filepath == path {
			return entry
		}
	}
	return nil
}

func (ide *IndexEntries) append(entry *IndexEntry) {
	ide.entries = append(ide.entries, entry)
}

func (ide *IndexEntries) updateOrInsert(entry *IndexEntry) {
	for i, e := range ide.entries {
		if e.Filepath == entry.Filepath {
			ide.entries[i] = entry
		}
	}
	ide.entries = append(ide.entries, entry)
	// sort.SliceStable(ide.entries, func(i, j int) bool {
	// 	return strings.Compare(ide.entries[i].Filepath, ide.entries[j].Filepath) < 0
	// })
}
func (ide *IndexEntries) Sort() {
	sort.SliceStable(ide.entries, func(i, j int) bool {
		return strings.Compare(ide.entries[i].Filepath, ide.entries[j].Filepath) < 0
	})
}

// func (ide *IndexEntries) insertEntries(entries []*IndexEntry) {
// 	// add, sort and update header
// 	ide.entries = append(ide.entries, entries...)
// 	// sort.SliceStable(ide.entries, func(i, j int) bool {
// 	// 	return strings.Compare(ide.entries[i].Filepath, ide.entries[j].Filepath) < 0
// 	// })
// }

func (ide *IndexEntries) remove(path string) bool {
	numOfEntries := len(ide.entries)
	for i, entry := range ide.entries {
		if entry.Filepath == path {
			if i < int(numOfEntries-1) {
				copy(ide.entries[i:], ide.entries[i+1:])
			}
			ide.entries = ide.entries[:numOfEntries-1]

			return true
		}
	}

	return false
}

type HandleIndexEntryFunc func(entry *IndexEntry)

func (ide *IndexEntries) foreach(fn HandleIndexEntryFunc) {
	for _, entry := range ide.entries {
		fn(entry)
	}
}

func (ide *IndexEntries) dumpIndexEntries(w io.Writer) {
	headerformat := "%-40s %-7s %8s %4s %4s %8s %8s  %-20s %-20s %-20s\n"
	fmt.Fprintf(w,
		headerformat,
		"Oid",
		"Mode",
		"Size",
		"Uid",
		"Gid",
		"Dev",
		"Ino",
		"Mtime",
		"Ctime",
		"Path",
	)

	lineFormat := "%20s %#o %8d %04d %4d %8d %8d %20v %20v  %s\n"
	ide.foreach(func(entry *IndexEntry) {
		fmt.Fprintf(w,
			lineFormat,
			entry.Oid,
			entry.Mode,
			entry.Size,
			entry.Uid,
			entry.Gid,
			entry.Dev,
			entry.Ino,
			entry.MTime.Format("2006-01-02T15:04:05"),
			entry.CTime.Format("2006-01-02T15:04:05"),
			entry.Filepath,
		)
	})
}
