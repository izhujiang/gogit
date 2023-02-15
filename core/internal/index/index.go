package index

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/internal/filemode"
)

// main data struct mapping to items in .git/index

const (
	sign_Index           = "DIRC"
	sign_ext_Tree        = "TREE"
	sign_ext_ResolveUndo = "REUC"
	sign_ext_Eoie        = "EOIE"
	sign_ext_link        = "link"
	sign_ext_UNTR        = "UNTR"
	sign_ext_FSMN        = "FSMN"
)

const (
	idx_version_2 = uint32(2)
	idx_version_3 = uint32(3)
	idx_version_4 = uint32(4)
)

const (
	NULL = 0x00

	maskFlagEntryExtended = 1 << 14
	maskFlagEntryStage    = uint16(0x3 << 12)
	maskFlagNameLength    = 0x0FFF

	maskExtflagSkipWorktree = uint16(1 << 14)
	maskExtflagIntentToAdd  = uint16(1 << 13)
	maskExtflagUnsed        = 0x1FFF
)

var (
	ErrNotOrInvalidIndexFile   = errors.New("This is not an valid index file.")
	ErrInvalidIndexFileVersion = errors.New("The version of this index file is not supported.")
	ErrInvalidTimestamp        = errors.New("Negative timestamps are not allowed")
	ErrCorruptedIndexFile      = errors.New("corrupted index file")
)

// Stage during merge
type Stage int32

const (
	// Merged is the default stage, fully merged
	Merged Stage = 1
	// AncestorMode is the base revision
	AncestorMode Stage = 1
	// OurMode is the first tree revision, ours
	OurMode Stage = 2
	// TheirMode is the second tree revision, theirs
	TheirMode Stage = 3
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
	Mode  filemode.FileMode
	Stage Stage
	Uid   uint32
	Gid   uint32

	// File size on-disk size from stat(2), truncated to 32-bit.
	Size uint32

	Skipworktree bool
	IntentToAdd  bool
}

func NewIndexEntry(oid common.Hash, mode filemode.FileMode, filepath string) *IndexEntry {
	ie := &IndexEntry{
		Mode:     mode,
		Filepath: filepath,
	}
	copy(ie.Oid[:], oid[:])

	return ie
}

// Tree contains pre-computed hashes for trees that can be derived from the
// index. It helps speed up tree object generation from index for a new commit.
type TreeCache struct {
	Entries []*TreeEntry
}

type TreeEntry struct {
	Name       string
	EntryCount int
	Subtrees   int
	Oid        common.Hash
}

func newTreeCache() *TreeCache {
	return &TreeCache{Entries: make([]*TreeEntry, 0)}

}

type ResolveUndoEntry struct {
	Path   string
	Stages map[Stage]common.Hash
}

type ResolveUndo struct {
	Entries []*ResolveUndoEntry
}

func newResolveUndo() *ResolveUndo {
	return &ResolveUndo{Entries: make([]*ResolveUndoEntry, 0)}

}

// unknown extention
type Extension struct {
	Signature []byte // If the first byte is 'A'..'Z' the extension is optional and can be ignored.
	Size      uint32 // 32-bit size of the extension
	Data      []byte // Extension data
}

type Index struct {
	Version           uint32
	NumberOfEntries   uint32
	Entries           []*IndexEntry
	TreeCache         *TreeCache
	ResolveUndo       *ResolveUndo
	unknownExtensions []*Extension
}

func New() *Index {
	return &Index{
		Version:           2,
		NumberOfEntries:   0,
		Entries:           make([]*IndexEntry, 0),
		TreeCache:         newTreeCache(),
		ResolveUndo:       newResolveUndo(),
		unknownExtensions: make([]*Extension, 0),
	}

}

func (idx *Index) RemoveAll() {
	idx.Entries = make([]*IndexEntry, 0)
	idx.NumberOfEntries = 0

}
func (idx *Index) InsertEntry(entry *IndexEntry) {
	// add, sort and update header
	idx.Entries = append(idx.Entries, entry)
	idx.NumberOfEntries += 1
	sort.SliceStable(idx.Entries, func(i, j int) bool {
		return strings.Compare(idx.Entries[i].Filepath, idx.Entries[j].Filepath) < 0
	})
}

func (idx *Index) InsertEntries(entries []*IndexEntry) {
	// add, sort and update header
	idx.Entries = append(idx.Entries, entries...)
	// idx.NumberOfEntries += uint32(len(entries))
	sort.SliceStable(idx.Entries, func(i, j int) bool {
		return strings.Compare(idx.Entries[i].Filepath, idx.Entries[j].Filepath) < 0
	})

	idx.NumberOfEntries = uint32(len(idx.Entries))
}

// Dump index file
func (idx *Index) Dump(w io.Writer) {
	fmt.Println("Index Entries:")
	idx.dumpIndexEntries(w)

	fmt.Fprintln(w)
	fmt.Println("Tree Entries:")
	// dump extentions
	idx.dumpTreeCache(w)
}

func (idx *Index) dumpIndexEntries(w io.Writer) {
	headerformat := "%-40s %-7s %8s %4s %4s %-8s %-8s  %-20s %-20s %-20s\n"
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
	for _, entry := range idx.Entries {
		// fmt.Fprintf(w, "%o %s %d \t%s\n", entry.Mode, entry.ObjectId, entry.StageNo(), string(entry.Path))
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
	}
}
func (idx *Index) dumpTreeCache(w io.Writer) {
	if idx.TreeCache != nil {
		headerformat := "%-40s %8s %8s  %-20s\n"
		fmt.Fprintf(w,
			headerformat,
			"Oid",
			"E_Count",
			"ST_Count",
			"Name",
		)

		lineFormat := "%20s %8d %8d  %s\n"
		for _, entry := range idx.TreeCache.Entries {

			fmt.Fprintf(w,
				lineFormat,
				entry.Oid,
				entry.EntryCount,
				entry.Subtrees,
				entry.Name,
			)
		}
	}

}
