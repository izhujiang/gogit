package index

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/izhujiang/gogit/common"
)

// Tree contains pre-computed hashes for trees that can be derived from the
// index. It helps speed up tree object generation from index for a new commit.

type CacheTreeEntry struct {
	Oid          common.Hash
	Name         string
	EntryCount   int
	SubtreeCount int
}

type CacheTree struct {
	entries   []*CacheTreeEntry
	hashtable map[string]*CacheTreeEntry
}

func newCacheTree() *CacheTree {
	return &CacheTree{
		entries: make([]*CacheTreeEntry, 0, 64),
		// TODO: change hashtable into more meaningful name, mapping from fullpath to *CacheTreeEntry
		hashtable: make(map[string]*CacheTreeEntry),
	}
}

func (c *CacheTree) reset() {
	c.entries = make([]*CacheTreeEntry, 0, 64)
	c.hashtable = make(map[string]*CacheTreeEntry)
}

func (c *CacheTree) append(te *CacheTreeEntry) {
	c.entries = append(c.entries, te)
}

type treeCacheEntryVisitHanlder func(*CacheTreeEntry)

// traval TreeCache by path, aa/bb/cc
func (c *CacheTree) visitByPath(path string, fn treeCacheEntryVisitHanlder) {
	path = filepath.Clean(path)
	for {
		if path == "." {
			path = ""
		}

		te, ok := c.hashtable[path]
		if ok {
			fn(te)
		}

		if path == "" {
			break
		}

		path = filepath.Dir(path)
	}
}

// Invalidate all TreeEntry in the path, for instance, InvalidatePath("aaa/bbb/ccc.txt") invalidate "", "aaa", "bbb"
func (c *CacheTree) invalidatePath(path string) {
	invalidateTreeEntryHanlder := func(entry *CacheTreeEntry) {
		entry.Oid = common.ZeroHash
		entry.EntryCount = -1
	}

	c.visitByPath(path, invalidateTreeEntryHanlder)
}

// only when TreeEntry.
func (c *CacheTree) findValidTreeCacheEntry(treePath string) (common.Hash, bool) {
	te, ok := c.hashtable[treePath]

	if ok && te.EntryCount >= 0 {
		return te.Oid, true
	} else {
		return common.ZeroHash, false
	}
}

func (c *CacheTree) buildTrees() {
	var fillHashtableEntry func(int, string)
	loopIndex := 0

	fillHashtableEntry = func(cur int, path string) {
		curItem := c.entries[cur]
		c.hashtable[path] = curItem

		for i := 0; i < curItem.SubtreeCount; i++ {
			loopIndex++
			fillHashtableEntry(loopIndex, filepath.Join(path, c.entries[loopIndex].Name))
		}
	}

	if len(c.entries) > 0 {
		fillHashtableEntry(0, "")
	}

	// for debug: sort and print
	// keys := make([]string, 0, len(c.hashtable))
	// for k := range c.hashtable {
	// 	keys = append(keys, k)
	// }
	// sort.Strings(keys)
	// for _, k := range keys {
	// 	fmt.Println(k, c.hashtable[k])
	// }

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
	for _, entry := range c.entries {
		fmt.Fprintf(w,
			lineFormat,
			entry.Oid,
			entry.EntryCount,
			entry.SubtreeCount,
			entry.Name,
		)
	}
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
