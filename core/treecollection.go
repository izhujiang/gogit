package core

import (
	"path/filepath"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/internal/filemode"
)

// mapping from fullpath to Tree object
type treeCollection map[string]*Tree

func (trees treeCollection) addFilePath(filePath string, mode filemode.FileMode, objType ObjectType, oid common.Hash) {
	filename := filepath.Base(filePath)
	// parent dir
	fp := filepath.Dir(filePath)
	for {
		tree, ok := trees[fp]
		if !ok {
			// fpName := filepath.Base(fp)
			tree = NewTree(common.ZeroHash, fp)
			trees[fp] = tree
		}
		te := newTreeEntry(mode, filename, oid)
		tree.entries.add(filename, te)

		mode = filemode.Dir
		oid = common.ZeroHash

		if fp == "." {
			break
		}

		filename = fp
		fp = filepath.Dir(filename)
	}
}

type WalkHandler func(t *Tree) error

// Depth-first Walk
func (trees treeCollection) DFWalk(fn WalkHandler) {
	root := trees["."]
	trees.walk(root, fn)
}

func (trees treeCollection) walk(tree *Tree, fn WalkHandler) {
	for _, entry := range tree.entries {
		if entry.Type == ObjectTypeTree {
			subtreePath := filepath.Join(tree.name, entry.Name)
			subtree := trees[subtreePath]
			trees.walk(subtree, fn)
			// oid might change
			copy(entry.Oid[:], subtree.oid[:])
		}
		// fmt.Printf("\t%s %s\t%s\t%s\n ", entry.Mode, entry.Type, entry.Oid, entry.Name)
	}
	fn(tree)
}
