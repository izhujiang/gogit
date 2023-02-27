package index

import (
	"path/filepath"
	"sort"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/object"
)

// mapping from fullpath to Tree object
type treeCollection map[string]*object.Tree

func (c treeCollection) addTreesByFilepath(filePath string, oid common.Hash, objType object.ObjectType, mode common.FileMode) {
	filename := filepath.Base(filePath)
	// parent dir
	treeName := filepath.Dir(filePath)
	if treeName == "." {
		treeName = ""
	}
	for {
		tree, ok := c[treeName]
		if !ok {
			// fpName := filepath.Base(fp)
			tree = object.NewTree(common.ZeroHash)
			c[treeName] = tree
		}
		te := object.NewTreeEntry(oid, filename, mode)
		tree.AddEntry(te)

		mode = common.Dir
		oid = common.ZeroHash

		if treeName == "" {
			break
		}

		filename = treeName
		treeName = filepath.Dir(filename)
		if treeName == "." {
			treeName = ""
		}
	}
}

type WalkHandler func(path string, t *object.Tree) error

func (c treeCollection) DFWalk(fn WalkHandler, preordering bool) {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}

	if preordering == true { // Top-down depth-first
		sort.Sort(sort.StringSlice(keys))
	} else { // Down-top depth-first
		sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	}

	for _, path := range keys {
		fn(path, c[path])
	}

}

// Depth-first Walk
// func (c treeCollection) DFWalk(fn WalkHandler, preordering bool) {
// 	var walk func(path string, fn WalkHandler)

// 	walk = func(path string, fn WalkHandler) {
// 		tree := c[path]

// 		if preordering == true {
// 			fn(path, tree)
// 		}

// 		tree.ForEachEntry(func(entry *object.TreeEntry) {
// 			if entry.Type == object.ObjectTypeTree {
// 				subtreePath := filepath.Join(path, entry.Name)
// 				walk(subtreePath, fn)
// 			}
// 			// fmt.Printf("\t%s %s\t%s\t%s\n ", entry.Mode, entry.Type, entry.Oid, entry.Name)
// 		})
// 		if preordering == false {
// 			fn(path, tree)
// 		}

// 	}
// 	root := ""
// 	walk(root, fn)
// }
