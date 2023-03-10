package object

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/izhujiang/gogit/common"
)

// mapping from fullpath to Tree object
type TreeCollection struct {
	root *Tree
	// size int
}

func NewTreeCollection() *TreeCollection {
	c := &TreeCollection{}

	return c
}
func (c *TreeCollection) InitWithRootId(oid common.Hash, path string) {
	t := c.MakeTreeAll(path)
	t.SetId(oid)
}

func (c *TreeCollection) InitWithRoot(r *Tree) {
	c.root = r
}

func (c *TreeCollection) Root() *Tree {
	return c.root
}

// MakeTreeAll creates a tree named path, along with any necessary parents, and returns nil, or else returns an error.
// Return the tree if it has alright existed, do nothing else.
func (c *TreeCollection) MakeTreeAll(path string) *Tree {
	path = filepath.Clean(path)
	path = filepath.ToSlash(path)

	// TODO: to deal with the exceptionnel path like "../../""
	if c.root == nil {
		c.root = NewTree(common.ZeroHash, "")
	}
	if path == "." || path == "" {
		return c.root
	}

	pathItems := strings.Split(path, "/")

	var makeTree func(*Tree, []string) *Tree
	makeTree = func(t *Tree, splitedPath []string) *Tree {
		subTree := t.Subtree(splitedPath[0])
		if subTree == nil {
			// te := NewTreeEntry(common.ZeroHash, filepath.Join(t.fullpath, subTreeName), common.Dir)
			te := NewTreeEntry(common.ZeroHash, splitedPath[0], common.Dir)
			t.UpdateOrAddEntry(te)
			subTree = te.(*Tree)
		}
		if len(splitedPath) > 1 {
			return makeTree(subTree, splitedPath[1:])
		} else {
			return subTree
		}
	}

	return makeTree(c.root, pathItems)

}

type WalkFunc func(*Tree)
type WalkWithPathFunc func(string, *Tree)
type WalkEntryWithPathFunc func(string, TreeEntry)

// Depth-first Walk
func (c *TreeCollection) DFWalk(fn WalkWithPathFunc, preordering bool) {
	var walk func(string, *Tree)

	walk = func(path string, t *Tree) {
		if preordering == true {
			fn(path, t)
		}

		t.ForEach(func(e TreeEntry) {
			if e.Type() == ObjectTypeTree {
				sub_path := filepath.Join(path, e.Name())
				walk(sub_path, e.(*Tree))
			}
			// fmt.Printf("\t%s %s\t%s\t%s\n ", entry.Mode, entry.Type, entry.Oid, entry.Name)
		})
		if preordering == false {
			fn(path, t)
		}

	}
	if c.root != nil {
		walk(c.root.name, c.root)
	}
}

func (c *TreeCollection) WalkByPath(path string, fn WalkFunc, preordering bool) {
	if c.root == nil {
		return
	}

	path = filepath.Clean(path)
	path = filepath.ToSlash(path)
	pathItems := strings.Split(path, "/")

	var walk func(*Tree, []string)

	walk = func(t *Tree, splitedPath []string) {
		if preordering == true {
			fn(t)
		}

		if len(splitedPath) > 0 {
			t.ForEach(func(e TreeEntry) {
				if e.Type() == ObjectTypeTree && splitedPath[0] == e.Name() {
					walk(e.(*Tree), pathItems[1:])
				}
				// fmt.Printf("\t%s %s\t%s\t%s\n ", entry.Mode, entry.Type, entry.Oid, entry.Name)
			})
		}
		if preordering == false {
			fn(t)
		}

	}

	if path == "." || path == "" {
		fn(c.root)
		return
	}

	walk(c.root, pathItems)
}

// assuming the treecollection hsa been expanded
func (c *TreeCollection) WalkByAlphabeticalOrder(fn WalkEntryWithPathFunc) {
	var walk func(string, *Tree)

	walk = func(path string, t *Tree) {
		t.ForEach(func(e TreeEntry) {
			sub_path := filepath.Join(path, e.Name())
			switch e.Type() {
			case ObjectTypeTree:
				walk(sub_path, e.(*Tree))

			case ObjectTypeBlob:
				fn(sub_path, e)

			default:
				panic("Not implemented")

			}
		})

	}
	if c.root != nil {
		walk(c.root.name, c.root)
	}
}
func (c *TreeCollection) Expand(fn WalkFunc) {
	if c.root == nil {
		return
	}

	tq := NewQueue()
	tq.Enqueue(c.root)
	for {
		t := tq.Dequeue()
		if t == nil {
			break
		}

		fn(t)
		// fmt.Println("t: ", t.Id(), t.Name())
		t.ForEach(func(e TreeEntry) {
			if e.Type() == ObjectTypeTree {
				// fmt.Println("enqueue t: ", e.Id(), e.Name())
				tq.Enqueue(e.(*Tree))
			}

		})
	}
}

func (c *TreeCollection) Merge(anothor *TreeCollection) {
	if anothor.root == nil {
		return
	}

	// o for originalTree, and n for newTree
	var mergeTrees func(*Tree, *Tree)

	mergeTrees = func(o *Tree, n *Tree) {
		changed := false
		n.ForEach(func(e TreeEntry) {
			switch e.Type() {
			case ObjectTypeBlob:
				o.UpdateOrAddEntry(e)
				changed = true
			case ObjectTypeTree:
				o_sub := o.Subtree(e.Name())
				if o_sub == nil {
					o.UpdateOrAddEntry(e)
					changed = true
				}
				n_sub := e.(*Tree)
				mergeTrees(o_sub, n_sub)
				if !changed && o_sub.oid == common.ZeroHash {
					changed = true
				}
			default:
				log.Fatal("Not valid tree entry", e.Id(), e.Name())
			}
		})
		if changed {
			o.oid = common.ZeroHash
		}
	}

	if c.root == nil {
		c.root = anothor.root
	} else {
		mergeTrees(c.root, anothor.root)
		return
	}

}

func (c *TreeCollection) Debug() {
	fmt.Println("Debug TreeCollection:")

	c.DFWalk(func(path string, t *Tree) {
		fmt.Println("path:", path, "id: ", t.Id(), "name:", t.Name())
		fmt.Println(t.Content())
	}, true)
}
