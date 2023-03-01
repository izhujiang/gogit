package object

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/izhujiang/gogit/common"
)

type entryOrder uint8

const (
	order_Ascending entryOrder = iota
	order_descending
)

// 100644 blob ec1871edcbfdc0d17ef498030e7ca676f291393d	LICENSE
//
//	type TreeEntry struct {
//		// Id of subtree or blob,  which this entry refer to
//		Oid  common.Hash
//		Name string
//		Type ObjectType
//		Mode common.FileMode
//	}

// Blob and Tree Object both implemente TreeEntry interface
type TreeEntry interface {
	Id() common.Hash
	Name() string
	Type() ObjectType
	Mode() common.FileMode
}

type TreeEntryCollection []TreeEntry

func newTreeEntryCollecion() TreeEntryCollection {
	tc := make([]TreeEntry, 0)
	return tc
}

func NewTreeEntry(id common.Hash, name string, mode common.FileMode) TreeEntry {
	// var oid common.Hash
	// copy(oid[:], refId[:])
	var te TreeEntry

	switch mode {
	case common.Dir:
		te = NewTree(id, name)
	case common.Regular:
		te = NewBlob(id, name, nil)
	default:
		panic("Not implemented.")
	}

	return te
}

// order by name of entry's name field
// 100644 blob 66fd13c903cac02eb9657cd53fb227823484401d	.gitignore
// 100644 blob ec1871edcbfdc0d17ef498030e7ca676f291393d	LICENSE
// 100644 blob db45ace0397638c3240baa7e31f1322aceeaec2f	Makefile
// 100644 blob efdd20273013935a0640c331219e5cb949f7fc2c	README.md
// 040000 tree 35c2f8041893ef999b84c24d13b5e0d3840770d8	api
// 040000 tree c3a4965c7c76750b5551b9f04899b58a29a3924b	cli
// 040000 tree 6dbba5e1f0f38b8c02d642327e188a0c6fcdcad0	core
// 100644 blob e22a1b515d156b422c21fceb379a239b00ddc4db	go.mod
// 100644 blob c6f955f3bb3aef39ac5d4ea2ca9674925a5c7c4e	go.sum
// 040000 tree c227a45a113be8f4482478d1f50a5eacde773371	plumbing
// 040000 tree 249696c6cb1ef790ccd683a6bcc704cdc5b97db5	porcelain
type Tree struct {
	// Hash ID
	oid common.Hash
	// fullpath string
	name string

	// TODO: make sure entries ordered by name
	// entries map[string]*TreeEntry
	entries TreeEntryCollection
}

func (t *Tree) Id() common.Hash {
	h := t.oid
	return h
}

func (t *Tree) SetId(oid common.Hash) {
	t.oid = oid
}

func (t *Tree) Name() string {
	// name := filepath.Base(t.fullpath)
	// if name == "." {
	// return ""
	// }
	return t.name
}

func (t *Tree) SetName(name string) {
	t.name = name
}

// func (t *Tree) SetPath(path string) {
// 	t.fullpath = filepath.Clean(path)
// }

// func (t *Tree) Path() string {
// 	return t.fullpath
// }

func (t *Tree) Type() ObjectType {
	return ObjectTypeTree
}

func (t *Tree) Mode() common.FileMode {
	return common.Dir
}

func NewTree(oid common.Hash, name string) *Tree {
	// path = filepath.Clean(path)
	// if path == "." {
	// path = ""
	// }

	return &Tree{
		oid: oid,
		// fullpath: path,
		name:    name,
		entries: newTreeEntryCollecion(),
	}
}
func EmptyTree() *Tree {
	return &Tree{
		entries: newTreeEntryCollecion(),
	}
}

type VisitTreeEntryFunc func(TreeEntry)

func (t *Tree) ForEach(fn VisitTreeEntryFunc) {
	for _, e := range t.entries {
		fn(e)
	}
}

func (t *Tree) Subtree(subtreeName string) *Tree {
	for _, e := range t.entries {
		if e.Type() == ObjectTypeTree && e.Name() == subtreeName {
			return e.(*Tree)
		}
	}

	return nil
}

func (t *Tree) UpdateOrAddEntry(entry TreeEntry) {
	for i, e := range t.entries {
		if e.Name() == entry.Name() {
			t.entries[i] = entry
			return
		}
	}
	t.entries = append(t.entries, entry)
}

func (t *Tree) SortEntries(order entryOrder) {
	entries := t.entries
	if order == order_Ascending {
		sort.SliceStable(entries, func(i, j int) bool {
			return strings.Compare(entries[i].Name(), entries[j].Name()) < 0
		})
	}
}

// TODO: output with format interface
func (t *Tree) Content() string {
	buf := &bytes.Buffer{}
	for _, e := range t.entries {
		// TODO: align the output
		fmt.Fprintf(buf, "%s %s %s\t%s\n", e.Mode(), e.Type(), e.Id(), e.Name())
	}

	return string(buf.Bytes())
}

// GitObject <==> Tree
func (t *Tree) FromGitObject(g *GitObject) {
	r := bytes.NewBuffer(g.content)
	entries := newTreeEntryCollecion()

	for {
		mode, err := r.ReadString(0x20)
		mode = strings.Trim(mode, " ")
		if err == io.EOF {
			break
		}
		name, _ := r.ReadBytes(0x00)
		fileName := string(name[:len(name)-1])
		var oid common.Hash
		_, _ = r.Read(oid[:])

		fm, _ := common.NewFileMode(mode)
		// entry := NewTreeEntry(oid, filepath.Join(t.fullpath, fileName), fm)
		entry := NewTreeEntry(oid, fileName, fm)

		entries = append(entries, entry)
	}

	t.oid = g.Hash()
	t.entries = entries
}

func (t *Tree) ToGitObject() *GitObject {
	w := &bytes.Buffer{}

	entries := t.entries
	for _, entry := range entries {
		mode := strings.TrimLeft(entry.Mode().String(), "0 ")
		w.WriteString(mode)
		w.WriteByte(0x20)
		w.WriteString(entry.Name())
		w.WriteByte(0x00)
		id := entry.Id()
		w.Write(id[:])
	}

	content := w.Bytes()

	g := &GitObject{
		objectType: ObjectTypeTree,
		size:       int64(len(content)),
		content:    content,
	}
	return g
}
