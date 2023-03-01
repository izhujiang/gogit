package index

import (
	"fmt"
	"io"

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
	// BUG: EntryCount, ASCII decimal number of entries in the index that is covered by the tree this entry represents (entry_count), include items in subtrees;
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
	object.TreeCollection
	// for encode and decode
	entries []*CacheTreeEntry
	// registers map[string]*CacheTreeEntry
}

func newCacheTree() *CacheTree {
	c := &CacheTree{
		entries: make([]*CacheTreeEntry, 0, cachetree_cap),
		// TODO: change hashtable into more meaningful name, mapping from fullpath to *CacheTreeEntry
		// registers: make(map[string]*CacheTreeEntry),
	}

	return c
}

func (c *CacheTree) reset() {
	c.entries = make([]*CacheTreeEntry, 0, cachetree_cap)
	c.buildTrees()

	// c.registers = make(map[string]*CacheTreeEntry)
}

func (c *CacheTree) buildTrees() {
	var newTreeFromCacheTreeEntry func(int, string) *object.Tree
	loopIndex := 0

	newTreeFromCacheTreeEntry = func(cur int, path string) *object.Tree {
		curItem := c.entries[cur]
		t := object.NewTree(curItem.Oid, path)
		// c.registers[path] = curItem

		for i := 0; i < curItem.SubtreeCount; i++ {
			loopIndex++
			sub_t := newTreeFromCacheTreeEntry(loopIndex, c.entries[loopIndex].Name)
			t.UpdateOrAddEntry(sub_t)
		}

		return t
	}

	if len(c.entries) > 0 {
		t := newTreeFromCacheTreeEntry(0, c.entries[0].Name)
		c.InitWithRoot(t)
	} else {
		c.InitWithRoot(nil)
	}
}

func (c *CacheTree) refreshCacheTreeEntries() {
	c.entries = make([]*CacheTreeEntry, 0, cachetree_cap)
	c.DFWalk(func(path string, t *object.Tree) {
		entryCount := 0
		subtreeCount := 0
		t.ForEach(func(e object.TreeEntry) {
			switch e.Mode() {
			case common.Dir:
				subtreeCount++
			case common.Regular:
				entryCount++
			}
		})

		if t.Id() == common.ZeroHash {
			entryCount = -1
		}

		cachetreeEntry := &CacheTreeEntry{
			Oid:          t.Id(),
			Name:         t.Name(),
			EntryCount:   entryCount,
			SubtreeCount: subtreeCount,
		}
		// fmt.Println("new cache tree entry: ", cachetreeEntry)
		c.entries = append(c.entries, cachetreeEntry)
	}, true)

	// sum up cache tree entry count bottom-up
	var updateCacheTreeEntryCount func(int) *CacheTreeEntry
	loopIndex := 0

	updateCacheTreeEntryCount = func(cur int) *CacheTreeEntry {
		curItem := c.entries[cur]
		totalEntryCount := 0

		for i := 0; i < curItem.SubtreeCount; i++ {
			loopIndex++
			subCacheTreeEntry := updateCacheTreeEntryCount(loopIndex)
			if subCacheTreeEntry.EntryCount == -1 {
				totalEntryCount = -1
			} else {
				totalEntryCount += subCacheTreeEntry.EntryCount
			}
		}

		if totalEntryCount != -1 || curItem.EntryCount != -1 {
			curItem.EntryCount += totalEntryCount
		} else {
			curItem.EntryCount = -1
		}

		return curItem
	}

	if len(c.entries) > 0 {
		updateCacheTreeEntryCount(0)
	}

}

func (c *CacheTree) foreach(fn CacheTreeEntryVisitFunc) {
	for _, e := range c.entries {
		fn(e)
	}
}

// Invalidate all TreeEntry in the path, for instance, InvalidatePath("aaa/bbb/ccc.txt") invalidate "", "aaa", "bbb"
func (c *CacheTree) invalidatePath(path string) {
	invalidateTreeEntryHanlder := func(t *object.Tree) {
		t.SetId(common.ZeroHash)
	}

	c.WalkByPath(path, invalidateTreeEntryHanlder, false)
}

// // only when TreeEntry.
// func (c *CacheTree) findValidTreeCacheEntry(path string) (common.Hash, bool) {
// 	// te, ok := c.registers[path]

// 	// if ok && te.EntryCount >= 0 {
// 	// 	return te.Oid, true
// 	// } else {
// 	return common.ZeroHash, false
// 	// }
// }

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
