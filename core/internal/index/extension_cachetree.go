package index

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/object"
)

// Tree contains pre-computed hashes for trees that can be derived from the
// index. It helps speed up tree object generation from index for a new commit.
type CacheTreeEntry struct {
	// Object name for the object that would result from writing this span of index as a tree
	Oid common.Hash
	// NUL-terminated path component (relative to its parent directory);
	Name string
	// EntryCount, ASCII decimal number of entries in the index that is covered by the tree this entry represents (entry_count), include items in subtrees;
	EntryCount int
	// ASCII decimal number that represents the number of subtrees this tree has;
	SubtreeCount int
}

type CacheTreeEntryVisitFunc func(*CacheTreeEntry)

const (
	cachetree_cap = 64
)

type CacheTree struct {
	// built-in trees
	object.TreeFs
	// for encode and decode
	cacheTreeEntries []*CacheTreeEntry
}

func newCacheTree() *CacheTree {
	c := &CacheTree{
		cacheTreeEntries: make([]*CacheTreeEntry, 0, cachetree_cap),
	}

	return c
}

func (c *CacheTree) reset() {
	c.cacheTreeEntries = make([]*CacheTreeEntry, 0, cachetree_cap)
	c.InitWithRoot(nil)
}

// load and trees from cacheTreeEntries
func (c *CacheTree) load() {
	// using items of cacheTreeEntries to build TreeFs
	var newTreeFromCacheTreeEntry func(int) *object.Tree
	loopIndex := 0

	newTreeFromCacheTreeEntry = func(cur int) *object.Tree {
		curItem := c.cacheTreeEntries[cur]
		t := object.NewTree(curItem.Oid)

		for i := 0; i < curItem.SubtreeCount; i++ {
			loopIndex++
			e_name := c.cacheTreeEntries[loopIndex].Name
			e_filemode := common.Dir

			sub_t := newTreeFromCacheTreeEntry(loopIndex)

			e := object.NewTreeEntry(sub_t.Id(), e_name, e_filemode)
			e.Pointer = sub_t

			t.Append(e)
		}

		return t
	}

	if len(c.cacheTreeEntries) > 0 {
		t := newTreeFromCacheTreeEntry(0)
		c.InitWithRoot(t)

	} else {
		c.InitWithRoot(nil)
	}
}

func (c *CacheTree) save() {
	c.cacheTreeEntries = make([]*CacheTreeEntry, 0, cachetree_cap)
	c.DFWalk(func(path string, t *object.Tree) error {
		entryCount := 0
		subtreeCount := 0
		t.ForEach(func(e *object.TreeEntry) error {
			switch e.Kind {
			case object.Kind_Tree:
				subtreeCount++
			case object.Kind_Blob:
				entryCount++
			}
			return nil
		})

		if t.Id() == common.ZeroHash {
			entryCount = -1
		}

		name := filepath.Base(path)
		if name == "." {
			name = ""
		}

		cachetreeEntry := &CacheTreeEntry{
			Oid:          t.Id(),
			Name:         name,
			EntryCount:   entryCount,
			SubtreeCount: subtreeCount,
		}
		// fmt.Println("new cache tree entry: ", cachetreeEntry)
		c.cacheTreeEntries = append(c.cacheTreeEntries, cachetreeEntry)

		return nil
	}, true)

	// sum up cache tree entry count bottom-up
	var updateCacheTreeEntryCount func(int) *CacheTreeEntry
	loopIndex := 0

	updateCacheTreeEntryCount = func(cur int) *CacheTreeEntry {
		curItem := c.cacheTreeEntries[cur]

		totalEntryCount := 0

		for i := 0; i < curItem.SubtreeCount; i++ {
			loopIndex++
			subCacheTreeEntry := updateCacheTreeEntryCount(loopIndex)
			if subCacheTreeEntry.EntryCount == -1 {
				totalEntryCount = -1
			} else {
				if totalEntryCount != -1 {
					totalEntryCount += subCacheTreeEntry.EntryCount
				}
			}
		}

		if curItem.EntryCount != -1 && totalEntryCount != -1 {
			curItem.EntryCount += totalEntryCount
		} else {
			curItem.EntryCount = -1
		}

		return curItem
	}

	if len(c.cacheTreeEntries) > 0 {
		updateCacheTreeEntryCount(0)
	}

}

func (c *CacheTree) foreach(fn CacheTreeEntryVisitFunc) {
	for _, e := range c.cacheTreeEntries {
		fn(e)
	}
}

// Invalidate all TreeEntry in the path, for instance, InvalidatePath("aaa/bbb/ccc.txt") invalidate "", "aaa", "bbb"
func (c *CacheTree) invalidatePath(path string) {
	invalidateTreeFn := func(t *object.Tree) error {
		t.ZeroId()
		return nil
	}

	c.WalkByPath(path, invalidateTreeFn, false)
}

// find the tree identified with path and invalidate it and it's subtrees
func (c *CacheTree) invalidatePathsWithPrefix(path string) {
	invalidateTreeFn := func(path string, t *object.Tree) error {
		t.ZeroId()
		return nil
	}

	c.DFWalkWithPrefix(path, invalidateTreeFn, false)
}

func (c *CacheTree) dump(w io.Writer) {
	headerformat := "%-40s %8s %8s  %-20s\n"
	fmt.Fprintf(w,
		headerformat,
		"Oid",
		"E_Count",
		"ST_Count",
		"Name",
	)

	lineFormat := "%20s %8d %8d  %s\n"

	c.foreach(func(e *CacheTreeEntry) {
		fmt.Fprintf(w,
			lineFormat,
			e.Oid,
			e.EntryCount,
			e.SubtreeCount,
			e.Name)
	})
}
