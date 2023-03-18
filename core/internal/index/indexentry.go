package index

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/izhujiang/gogit/common"
)

type IndexEntry struct {
	oid      common.Hash
	filepath string

	fileinfo

	stage        Stage
	skipworktree bool
	intentToAdd  bool
}

func NewIndexEntry(oid common.Hash, mode common.FileMode, fpath string) *IndexEntry {
	e := &IndexEntry{
		oid: oid,
		fileinfo: fileinfo{
			name: filepath.Base(fpath),
			mode: mode,
		},
		filepath: fpath,
	}
	return e
}

func NewIndexEntryWithFileInfo(oid common.Hash, mode common.FileMode, fpath string, fi os.FileInfo) *IndexEntry {
	e := &IndexEntry{
		oid: oid,
		fileinfo: fileinfo{
			name: filepath.Base(fpath),
			mode: mode,
		},
		filepath: fpath,
	}

	stat := fi.Sys().(*syscall.Stat_t)
	e.mode = common.FileMode(stat.Mode)
	e.cTime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	e.mTime = fi.ModTime()
	e.dev = uint32(stat.Dev)
	e.ino = uint32(stat.Ino)
	e.uid = stat.Gid
	e.gid = stat.Gid
	e.size = uint32(stat.Size)
	return e
}
func (e *IndexEntry) UpdateWithFileInfo(fi os.FileInfo) {
	stat := fi.Sys().(*syscall.Stat_t)
	e.mode = common.FileMode(stat.Mode)
	e.cTime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	e.mTime = fi.ModTime()
	e.dev = uint32(stat.Dev)
	e.ino = uint32(stat.Ino)
	e.uid = stat.Gid
	e.gid = stat.Gid
	e.size = uint32(stat.Size)
}

type IndexEntries struct {
	entries []*IndexEntry
}

func (ide *IndexEntries) LsIndexEntries(w io.Writer, withDetail bool) {
	if withDetail {
		for _, e := range ide.entries {
			fmt.Fprintf(w, "%o %s %d \t%s\n", e.mode, e.oid, e.stage, e.filepath)
		}
	} else {
		for _, e := range ide.entries {
			fmt.Fprintln(w, e.filepath)
		}

	}
}

func (ide *IndexEntries) size() int {
	return len(ide.entries)
}
func (ide *IndexEntries) reset() {
	ide.entries = make([]*IndexEntry, 0, 256)
}

func (ide *IndexEntries) find(path string) *IndexEntry {
	for _, entry := range ide.entries {
		if entry.filepath == path {
			return entry
		}
	}
	return nil
}

// func (ide *IndexEntries) append(entry *IndexEntry) {
// 	ide.entries = append(ide.entries, entry)
// }

func (ide *IndexEntries) updateOrAppend(entry *IndexEntry) {
	for i, e := range ide.entries {
		if e.filepath == entry.filepath {
			ide.entries[i] = entry
			return
		}
	}
	ide.entries = append(ide.entries, entry)
	// sort.SliceStable(ide.entries, func(i, j int) bool {
	// 	return strings.Compare(ide.entries[i].Filepath, ide.entries[j].Filepath) < 0
	// })
}

func (ide *IndexEntries) append(entry *IndexEntry) {
	ide.entries = append(ide.entries, entry)
}

func (ide *IndexEntries) Sort() {
	sort.SliceStable(ide.entries, func(i, j int) bool {
		return strings.Compare(ide.entries[i].filepath, ide.entries[j].filepath) < 0
	})
}

func (ide *IndexEntries) remove(path string) bool {
	numOfEntries := len(ide.entries)
	for i, entry := range ide.entries {
		if entry.filepath == path {
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
			entry.oid,
			entry.mode,
			entry.size,
			entry.uid,
			entry.gid,
			entry.dev,
			entry.ino,
			entry.mTime.Format("2006-01-02T15:04:05"),
			entry.cTime.Format("2006-01-02T15:04:05"),
			entry.filepath,
		)
	})
}
