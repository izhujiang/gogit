package object

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/izhujiang/gogit/common"
)

// mapping from fullpath to Tree object
type TreeFs struct {
	root *Tree
	// size int
}

func EmptyTreeFs() *TreeFs {
	fs := &TreeFs{}

	return fs
}

func NewTreeFs(root *Tree) *TreeFs {
	fs := &TreeFs{
		root,
	}

	return fs
}

func (fs *TreeFs) InitWithRoot(r *Tree) {
	fs.root = r
}

func (fs *TreeFs) Root() *Tree {
	return fs.root
}

// MakeTreeAll creates a tree named path, along with any necessary parents, and returns nil, or else returns an error.
// Return the tree if it has alright existed, do nothing else.
func (fs *TreeFs) MakeTreeAll(path string) *Tree {

	// TODO: to deal with the exceptionnel path like "../../""
	if fs.root == nil {
		fs.root = NewTree(common.ZeroHash)
	}
	if path == "." || path == "" {
		return fs.root
	}

	pathItems := strings.Split(path, "/")

	var makeTree func(*Tree, []string) *Tree
	makeTree = func(t *Tree, splitedPath []string) *Tree {
		te := t.Subtree(splitedPath[0])
		if te == nil {
			// te := NewTreeEntry(common.ZeroHash, filepath.Join(t.fullpath, subTreeName), common.Dir)
			te = NewTreeEntry(common.ZeroHash, splitedPath[0], common.Dir)
			te.Pointer = EmptyTree()
			t.Append(te)
		}

		if len(splitedPath) > 1 {
			return makeTree(te.Pointer.(*Tree), splitedPath[1:])
		} else {
			return te.Pointer.(*Tree)
		}
	}

	return makeTree(fs.root, pathItems)

}

type WalkFunc func(*Tree) error
type WalkWithPathFunc func(string, *Tree) error

// Depth-first Walk, travel all trees
func (fs *TreeFs) DFWalk(fn WalkWithPathFunc, preordering bool) {
	fs.DFWalkWithPrefix("", fn, preordering)
}

func (fs *TreeFs) DFWalkWithPrefix(prefix string, fn WalkWithPathFunc, preordering bool) {
	if fs.root == nil {
		return
	}

	var walk func(string, *Tree)

	walk = func(path string, t *Tree) {
		if preordering == true {
			err := fn(path, t)

			if err == filepath.SkipDir {
				return
			}
		}

		t.ForEach(func(e *TreeEntry) error {
			if e.Kind == Kind_Tree {
				sub_path := filepath.Join(path, e.Name)
				walk(sub_path, e.Pointer.(*Tree))
			}
			// fmt.Printf("\t%s %s\t%s\t%s\n ", entry.Mode, entry.Type, entry.Oid, entry.Name)
			return nil
		})

		if preordering == false {
			err := fn(path, t)
			if err == filepath.SkipDir {
				return
			}
		}
	}

	if prefix == "" {
		walk("", fs.root)
	} else {
		t := fs.Find(prefix)
		if t != nil {
			walk(prefix, t)
		}
	}
}

func (fs *TreeFs) Find(path string) *Tree {
	if fs.root == nil {
		return nil
	}

	path = filepath.Clean(path)
	path = filepath.ToSlash(path)
	path = strings.TrimLeft(path, "/")
	pathItems := strings.Split(path, "/")

	t := fs.root
	for _, subpath := range pathItems {
		var sub_t *Tree
		for _, e := range t.entries {
			if e.Kind == Kind_Tree && e.Name == subpath {
				sub_t = e.Pointer.(*Tree)
				break
			}
		}

		if sub_t == nil {
			return nil
		}

		t = sub_t
	}

	return t

}

// WalkbyPath, travel all trees that along with the path
func (fs *TreeFs) WalkByPath(path string, fn WalkFunc, preordering bool) {
	if fs.root == nil {
		return
	}

	path = filepath.Clean(path)
	path = filepath.ToSlash(path)
	path = strings.TrimLeft(path, "/")
	pathItems := strings.Split(path, "/")

	var walk func(*Tree, []string)

	walk = func(t *Tree, splitedSubPaths []string) {
		if preordering == true {
			fn(t)
		}

		if len(splitedSubPaths) > 0 {
			t.ForEach(func(e *TreeEntry) error {
				if e.Kind == Kind_Tree && splitedSubPaths[0] == e.Name {
					walk(e.Pointer.(*Tree), pathItems[1:])
				}
				// fmt.Printf("\t%s %s\t%s\t%s\n ", entry.Mode, entry.Type, entry.Oid, entry.Name)
				return nil
			})
		}
		if preordering == false {
			fn(t)
		}

	}

	// no subpath
	if path == "." || path == "" {
		fn(fs.root)
		return
	}

	walk(fs.root, pathItems)
}

func (fs *TreeFs) Merge(anothor *TreeFs) {
	if anothor.root == nil {
		return
	}

	// o for originalTree, and n for newTree
	var mergeTrees func(*Tree, *Tree)

	mergeTrees = func(o *Tree, n *Tree) {
		changed := false
		n.ForEach(func(e *TreeEntry) error {
			switch e.Kind {
			case Kind_Blob:
				if o_e := o.Find(e.Name); o_e != nil {
					if o_e.Oid != e.Oid {
						return filepath.SkipDir
					}
				} else {
					o.Append(e)
					changed = true
				}
			case Kind_Tree:
				o_sub := o.Subtree(e.Name)
				if o_sub == nil {
					o.Append(e)
					changed = true
				} else {
					n_sub_t := e.Pointer.(*Tree)
					o_sub_t := o_sub.Pointer.(*Tree)
					mergeTrees(o_sub_t, n_sub_t)

					if o_sub_t.oid == common.ZeroHash {
						o_sub.Oid = common.ZeroHash
						changed = true
					}
				}
			default:
				log.Fatal("Not valid tree entry", e.Oid, e.Name)
			}
			return nil
		})
		if changed {
			o.oid = common.ZeroHash
		}
	}

	if fs.root == nil {
		fs.root = anothor.root
	} else {
		mergeTrees(fs.root, anothor.root)

		// remove empty subtrees and udpate t's tree entries
		fs.DFWalk(func(path string, t *Tree) error {
			t.RegularizeEntries()
			return nil
		}, false)

		return
	}
}

func (c *TreeFs) Debug() {
	fmt.Println("Debug TreeFs:")

	c.DFWalk(func(path string, t *Tree) error {
		fmt.Println("id:", t.Id())

		fmt.Println("content of tree:", t.Content())
		return nil
	}, true)
}
