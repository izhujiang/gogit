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
func (e *IndexEntry) Update(oid common.Hash, fi os.FileInfo) {
	e.oid = oid
	if fi != nil {
		stat := fi.Sys().(*syscall.Stat_t)
		e.mode = common.FileMode(stat.Mode)
		e.cTime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
		e.mTime = fi.ModTime()
		e.dev = uint32(stat.Dev)
		e.ino = uint32(stat.Ino)
		e.uid = stat.Gid
		e.gid = stat.Gid
		e.size = uint32(stat.Size)
	} else {
		// e.mode = common.FileMode(stat.Mode)
		e.cTime = time.Unix(int64(0), int64(0))
		e.mTime = time.Unix(int64(0), int64(0))
		e.dev = 0
		e.ino = 0
		e.uid = 0
		e.gid = 0
		e.size = 0
	}
}

type IndexEntries struct {
	entries []*IndexEntry
}

func (ide *IndexEntries) ListIndex(w io.Writer, withDetail bool) {
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

func (ide *IndexEntries) Find(path string) *IndexEntry {
	for _, e := range ide.entries {
		if e.filepath == path {
			return e
		}
	}
	return nil
}

func (ide *IndexEntries) Sort() {
	sort.SliceStable(ide.entries, func(i, j int) bool {
		return strings.Compare(ide.entries[i].filepath, ide.entries[j].filepath) < 0
	})
}

func (ide *IndexEntries) size() int {
	return len(ide.entries)
}

func (ide *IndexEntries) reset() {
	ide.entries = make([]*IndexEntry, 0, 256)
}

func (ide *IndexEntries) append(entry *IndexEntry) {
	ide.entries = append(ide.entries, entry)
}
func (ide *IndexEntries) remove(path string) bool {
	numOfEntries := len(ide.entries)
	for i, e := range ide.entries {
		if e.filepath == path {
			fmt.Printf("rm '%s'\n", path)
			if i < int(numOfEntries-1) {
				copy(ide.entries[i:], ide.entries[i+1:])
			}
			ide.entries = ide.entries[:numOfEntries-1]

			return true
		}
	}

	return false
}

func (ide *IndexEntries) removeWithPrefix(path string) bool {
	numOfEntries := len(ide.entries)
	entries := make([]*IndexEntry, numOfEntries)

	i := 0
	for _, e := range ide.entries {
		if !strings.HasPrefix(e.filepath, path) {
			entries[i] = e
			i++
		} else {
			fmt.Printf("rm '%s'\n", e.filepath)
		}
	}

	if i == numOfEntries {
		return false
	} else {
		ide.entries = entries[:i]
		return true
	}
}

func (ide *IndexEntries) dump(w io.Writer) {
	headerformat := "%-40s %8s %8s %4s %4s %8s %8s  %-20s %-20s %-20s\n"
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

	lineFormat := "%20s %8o %8d %4d %4d %8d %8d %20v %20v  %s\n"
	for _, e := range ide.entries {
		fmt.Fprintf(w,
			lineFormat,
			e.oid,
			e.mode,
			e.size,
			e.uid,
			e.gid,
			e.dev,
			e.ino,
			e.mTime.Format("2006-01-02T15:04:05"),
			e.cTime.Format("2006-01-02T15:04:05"),
			e.filepath,
		)
	}
}
