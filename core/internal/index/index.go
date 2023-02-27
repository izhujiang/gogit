package index

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/object"
)

// main data struct mapping to items in .git/index
// git has a file called the index that it uses to keep track of the file changes over the three areas: working directory, staging area, and repository.

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
	maskFlagEntryExtended = 1 << 14
	maskFlagEntryStage    = uint16(0x3 << 12)
	maskFlagNameLength    = 0x0FFF

	maskExtflagSkipWorktree = uint16(1 << 14)
	maskExtflagIntentToAdd  = uint16(1 << 13)
	maskExtflagUnsed        = 0x1FFF
)

const (
	sep_NULL    byte = 0x00
	sep_SPACE   byte = 0x20
	sep_NEWLINE byte = 0x0A
)

var (
	ErrNotOrInvalidIndexFile   = errors.New("This is not an valid index file.")
	ErrInvalidIndexFileVersion = errors.New("The version of this index file is not supported.")
	ErrInvalidTimestamp        = errors.New("Negative timestamps are not allowed")
	ErrCorruptedIndexFile      = errors.New("corrupted index file")
)

var (
	ErrIndexEntryNotExists = errors.New("There is no such index entry exists.")
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

type Index struct {
	version              uint32
	numberOfIndexEntries uint32
	// Entries           []*IndexEntry
	IndexEntries
	cacheTree         *CacheTree
	unsolveUndo       *ResolveUndo
	unknownExtensions []*Extension
}

func newIndex() *Index {
	idx := &Index{
		version:              2,
		numberOfIndexEntries: 0,
		IndexEntries:         IndexEntries{},
		// TreeCache:         newTreeCache(),
		// ResolveUndo:       newResolveUndo(),
		unknownExtensions: make([]*Extension, 0),
	}
	idx.reset()

	return idx
}

func (idx *Index) Reset() {
	idx.reset()
	idx.numberOfIndexEntries = 0

	if idx.cacheTree != nil {
		idx.cacheTree.reset()
	}
}

func (idx *Index) FindIndexEntry(path string) (*IndexEntry, error) {
	return idx.find(path)
}

func (idx *Index) ForeachIndexEntry(fn HandleIndexEntryFunc) {
	idx.foreach(fn)
}

func (idx *Index) InsertIndexEntry(entry *IndexEntry) {
	idx.insert(entry)
	idx.numberOfIndexEntries = uint32(idx.size())

	if idx.cacheTree != nil {
		idx.cacheTree.invalidatePath(entry.Filepath)
	}
}
func (idx *Index) InsertEntries(entries []*IndexEntry) {
	idx.insertEntries(entries)
	idx.numberOfIndexEntries = uint32(idx.size())

	if idx.cacheTree != nil {
		idx.foreach(func(e *IndexEntry) {
			idx.cacheTree.invalidatePath(e.Filepath)
		})
	}
}

func (idx *Index) RemoveIndexEntry(path string) {
	idx.remove(path)
	idx.numberOfIndexEntries = uint32(idx.size())

	if idx.cacheTree != nil {
		idx.cacheTree.invalidatePath(path)
	}
}

func (idx *Index) NewCacheTree() {
	idx.cacheTree = newCacheTree()
}

func (idx *Index) AppendCacheTreeEntry(entry *CacheTreeEntry) {
	if idx.cacheTree != nil {
		idx.cacheTree.append(entry)
	}
}

func (idx *Index) BuildCacheTree() {
	if idx.cacheTree != nil {
		idx.cacheTree.buildTrees()
	}
}

func (idx *Index) InvalidatePathInCacheTree(path string) {
	if idx.cacheTree != nil {
		idx.cacheTree.invalidatePath(path)
	}
}
func (idx *Index) FindValidTreeCacheEntry(path string) (common.Hash, bool) {
	if idx.cacheTree != nil {
		return idx.cacheTree.findValidTreeCacheEntry(path)
	}

	return common.ZeroHash, false
}

// using files in the index entries to build trees
type SaveTreeFunc func(t *object.Tree)

func (idx *Index) WriteTree(fn SaveTreeFunc) (common.Hash, error) {
	trees := treeCollection{}

	// setup treeCollection
	idx.foreach(func(e *IndexEntry) {
		// fullfill treeCollection
		// fmt.Printf("%o %s %d \t%s\n", e.Mode, e.Oid, e.Stage, e.Filepath)
		trees.addTreesByFilepath(
			e.Filepath,
			e.Oid,
			object.ObjectTypeBlob,
			e.Mode)
	})

	// Hash the tree according to new entries and save to repository
	updateTree := func(path string, t *object.Tree) error {
		// fmt.Println("before updating current tree: ", path, t.Id())
		if idx.cacheTree != nil {
			oid, found := idx.cacheTree.findValidTreeCacheEntry(path)
			if found {
				t.SetId(oid)
				return nil
			}
		}

		// has not found in TreeCache
		t.ForEachEntry(func(e *object.TreeEntry) {
			if e.Type == object.ObjectTypeTree {
				subpath := filepath.Join(path, e.Name)
				subTree := trees[subpath]
				e.Oid = subTree.Id()
			}
		})
		fn(t)

		// fmt.Println("after updating current tree: ", path, t.Id())
		// t.ShowContent(os.Stdout)
		return nil
	}

	trees.DFWalk(updateTree, false)

	// build new CacheTree
	idx.cacheTree = newCacheTree()
	generateTreeCacheEntry := func(path string, t *object.Tree) error {
		entryCount := 0
		subtreeCount := 0
		t.ForEachEntry(func(e *object.TreeEntry) {
			if e.Mode.IsFile() {
				entryCount++
			} else {
				subtreeCount++
			}
		})

		cachetreeEntry := &CacheTreeEntry{
			Oid:          t.Id(),
			Name:         path,
			EntryCount:   entryCount,
			SubtreeCount: subtreeCount,
		}
		idx.AppendCacheTreeEntry(cachetreeEntry)

		return nil
	}
	trees.DFWalk(generateTreeCacheEntry, true)

	return trees[""].Id(), nil
}

// Dump index file
func (idx *Index) Dump(w io.Writer) {
	fmt.Println("Index Entries:")
	idx.dumpIndexEntries(w)

	fmt.Fprintln(w)

	if idx.cacheTree != nil {
		fmt.Println("Tree Entries:")
		idx.cacheTree.dump(w)
	}
}

func LoadIndex(path string) *Index {
	idx := newIndex()

	f, err := os.Open(path)
	// *PathError
	if err != nil {
		return idx
	}
	defer f.Close()

	decoder := NewIndexDecoder(f)
	decoder.Decode(idx)

	idx.BuildCacheTree()

	return idx
}

func (idx *Index) SaveIndex(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := NewIndexEncoder(f)
	encoder.Encode(idx)
	return nil
}
