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
	version uint32
	// numberOfIndexEntries uint32
	// Entries           []*IndexEntry
	IndexEntries
	CacheTree
	unsolveUndo       *ResolveUndo
	unknownExtensions []*Extension
}

func (idx *Index) Load(path string) {
	f, err := os.Open(path)
	if err != nil {
		idx.version = idx_version_2
		idx.IndexEntries.reset()
		return
	}
	defer f.Close()

	decoder := NewIndexDecoder(f)
	decoder.Decode(idx)

	// load CacheTree and fully fill with index entries
	idx.CacheTree.load()

	// fill cacheTree with index entries
	if idx.CacheTree.Root() != nil {
		idx.Foreach(func(e *IndexEntry) {
			dir := common.DirOfFilePath(e.filepath)

			t := idx.CacheTree.Find(dir)
			if t != nil {
				base := filepath.Base(e.filepath)
				te := object.NewTreeEntry(e.oid, base, e.mode)
				t.Append(te)
			}
		})

		// build cacnheTree
		// idx.CacheTree.DFWalk(func(path string, t *object.Tree) error {
		// 	t.Sort()

		// 	return nil
		// }, false)
	}
}

func (idx *Index) Save(path string) error {
	idx.CacheTree.updateCacheTreeEntries()

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
	idx.IndexEntries.reset()
	// idx.numberOfIndexEntries = 0

	idx.CacheTree.reset()
}

type WalkIndexEntryFunc func(*IndexEntry)

func (idx *Index) Foreach(fn WalkIndexEntryFunc) {
	for _, e := range idx.entries {
		fn(e)
	}
}

func (idx *Index) Append(e *IndexEntry) {
	idx.append(e)
	// idx.numberOfIndexEntries = uint32(idx.size())

	idx.CacheTree.invalidatePath(filepath.Dir(e.filepath))
}

func (idx *Index) Update(e *IndexEntry, oid common.Hash, fi os.FileInfo) {
	e.Update(oid, fi)

	idx.CacheTree.invalidatePath(filepath.Dir(e.filepath))

}
func (idx *Index) Remove(path string, recursive bool) {
	if recursive == true {
		removed := idx.removeWithPrefix(path)
		if removed {
			idx.CacheTree.invalidatePath(filepath.Dir(path))
			idx.CacheTree.invalidatePathsWithPrefix(path)
		}

	} else {
		removed := idx.remove(path)

		if removed {
			dir := filepath.Dir(path)
			idx.CacheTree.invalidatePath(dir)
		}

	}

	// idx.numberOfIndexEntries = uint32(idx.size())
}

// using files in the index entries to build trees
func (idx *Index) UpdateCacheTree() {
	// cacheTree is already valid, do nothing
	root := idx.CacheTree.Root()

	if root == nil || root.Id() == common.ZeroHash {
		// path --> *Tree map, cache Tree has been created
		treeMap := make(map[string]*object.Tree)
		idx.Foreach(func(e *IndexEntry) {
			// TODO: to skip if the trees alone with e.filepath have non-zero id, which means the trees ware not invalid

			dir := common.DirOfFilePath(e.filepath)
			filename := filepath.Base(e.filepath)

			// make tree or return the existed tree
			t, ok := treeMap[dir]
			if !ok {
				t = idx.CacheTree.MakeTreeAll(dir)
				treeMap[dir] = t
			}
			te := t.Find(filename)
			if te == nil {
				t.Append(object.NewTreeEntry(e.oid, filename, e.mode))
			}
		})

	}
}

// ReadTrees read all the expanded trees from TreeFs  and put them into cacheTree
func (idx *Index) ReadTrees(fs *object.TreeFs) error {
	// add into index entries with all files in trees
	errMsg := &bytes.Buffer{}
	fs.DFWalk(func(path string, t *object.Tree) error {
		t.ForEach(func(e *object.TreeEntry) error {
			if e.Kind == object.Kind_Blob {
				fullfilepath := filepath.Join(path, e.Name)
				idxEntry := idx.Find(fullfilepath)
				if idxEntry == nil {
					idxEntry = NewIndexEntry(e.Oid, e.Filemode, fullfilepath)
					idx.append(idxEntry)
				} else {
					fmt.Fprintf(errMsg, "Entry '%s' overlaps with '%s'.  Cannot bind.\n", fullfilepath, fullfilepath)
				}
			}

			return nil
		})

		return nil
	}, true)
	// idx.numberOfIndexEntries = uint32(idx.size())

	if errMsg.Len() > 0 {
		return errors.New(string(errMsg.Bytes()))
	}
	idx.Sort()

	idx.CacheTree.Merge(fs)

	return nil
}

// Dump index file
func (idx *Index) Dump(w io.Writer) {
	fmt.Println("Index Entries:")
	idx.IndexEntries.dump(w)

	fmt.Fprintln(w)

	idx.CacheTree.dump(w)
}
