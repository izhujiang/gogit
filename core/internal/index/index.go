package index

import (
	"bytes"
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
	ErrCorruptedIndexFile      = errors.New("Corrupted index file")
	ErrCacheTreeAllValid       = errors.New("All trees in the cache are already valid and need no more regenerate cachetrees")
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

	return idx
}

func (idx *Index) SaveIndex(path string) error {
	if idx.cacheTree != nil {
		idx.cacheTree.refreshCacheTreeEntries()
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := NewIndexEncoder(f)
	encoder.Encode(idx)
	return nil
}

func (idx *Index) Reset() {
	idx.reset()
	idx.numberOfIndexEntries = 0

	if idx.cacheTree != nil {
		idx.cacheTree.reset()
	}
}

func (idx *Index) ForeachIndexEntry(fn HandleIndexEntryFunc) {
	idx.foreach(fn)
}

func (idx *Index) FindIndexEntry(path string) *IndexEntry {
	return idx.find(path)
}

// func (idx *Index) UpdateOrInsertIndexEntry(entry *IndexEntry, updateCacheTree bool) {
func (idx *Index) UpdateOrInsertIndexEntry(entry *IndexEntry) {
	idx.updateOrAppend(entry)
	idx.numberOfIndexEntries = uint32(idx.size())

	if idx.cacheTree != nil {
		idx.cacheTree.invalidatePath(entry.filepath)
	}
}

// func (idx *Index) RemoveIndexEntry(path string, updateCacheTree bool) {
func (idx *Index) RemoveIndexEntry(path string) {
	idx.remove(path)
	idx.numberOfIndexEntries = uint32(idx.size())

	if idx.cacheTree != nil {
		idx.cacheTree.invalidatePath(path)
	}
}

// using files in the index entries to build trees
func (idx *Index) WriteTree(saveTreeFn object.WalkFunc) (common.Hash, error) {
	// cacheTree is already valid, do nothing
	if idx.cacheTree != nil && idx.cacheTree.Root().Id() != common.ZeroHash {
		return idx.cacheTree.Root().Id(), ErrCacheTreeAllValid
	}

	if idx.cacheTree == nil {
		idx.cacheTree = newCacheTree()
	}

	// idx.cacheTree.Debug()

	idx.foreach(func(e *IndexEntry) {
		dir := filepath.Dir(e.filepath)
		filename := filepath.Base(e.filepath)

		// make tree or return the existed tree
		t := idx.cacheTree.MakeTreeAll(dir)
		t.UpdateOrAddEntry(object.NewTreeEntry(e.oid, filename, common.Regular))
	})

	// Hash the tree with hash-code from cacheTree or fn(SaveTreeFunc)
	idx.cacheTree.DFWalk(func(path string, t *object.Tree) error {
		if t.Id() == common.ZeroHash {
			t.Sort()
			saveTreeFn(t)
		}

		return nil
	}, false)

	// idx.cacheTree.refreshCacheTreeEntries()
	return idx.cacheTree.Root().Id(), nil
}

// ReadTrees read all the expanded trees from TreeCollection and put them into cacheTree
func (idx *Index) ReadTrees(trees *object.TreeFs, saveTreeFn object.WalkFunc) error {
	// add index entries
	errMsg := &bytes.Buffer{}
	trees.DFWalk(func(path string, t *object.Tree) error {
		t.ForEach(func(e object.TreeEntry) error {
			if e.Kind() == object.Kind_Blob {
				fullfilepath := filepath.Join(path, e.Name())
				idxEntry := idx.find(fullfilepath)
				if idxEntry == nil {
					idxEntry = NewIndexEntry(e.Id(), e.Type(), fullfilepath)
					idx.append(idxEntry)
				} else {
					fmt.Fprintf(errMsg, "Entry '%s' overlaps with '%s'.  Cannot bind.\n", fullfilepath, fullfilepath)
				}
			}

			return nil
		})

		return nil
	}, true)
	idx.numberOfIndexEntries = uint32(idx.size())

	if errMsg.Len() > 0 {
		return errors.New(string(errMsg.Bytes()))
	}
	idx.Sort()

	// add new cachetree entries from the read trees
	if idx.cacheTree == nil {
		idx.cacheTree = newCacheTree()
		// 	// idx.cacheTree.buildTreeFs()
	}

	// // setup treeFs
	// idx.foreach(func(e *IndexEntry) {
	// 	dir := filepath.Dir(e.filepath)
	// 	filename := filepath.Base(e.filepath)

	// 	// make tree or return the existed tree
	// 	t := idx.cacheTree.MakeTreeAll(dir)
	// 	t.UpdateOrAddEntry(object.NewTreeEntry(e.oid, filename, common.Regular))
	// })

	fmt.Println("before mrege cachetree")
	idx.cacheTree.Debug()

	fmt.Println("before mrege trees:")
	trees.Debug()
	idx.cacheTree.Merge(trees)

	fmt.Println("after mrege cachetree")
	idx.cacheTree.Debug()

	// Hash the tree with hash-code from cacheTree or fn(SaveTreeFunc)
	idx.cacheTree.DFWalk(func(path string, t *object.Tree) error {
		// fmt.Println("before updating current tree: ", path, t.Id(), t.Name())
		if t.Id() == common.ZeroHash {
			t.Sort()
			saveTreeFn(t)
			// fmt.Println("after updating current tree: ", t.Id(), t.Name())
		} else {
			// just skip
		}

		// fmt.Println("current tree: ", t.Id(), t.Name())
		// fmt.Println(t.Content())

		return nil
	}, false)

	return nil
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

// ----------------------------------------------
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
