package object

import (
	"fmt"
	"io"
	"sort"

	"github.com/izhujiang/gogit/common"
)

// 100644 blob ec1871edcbfdc0d17ef498030e7ca676f291393d	LICENSE
type TreeEntry struct {
	// Id of subtree or blob,  which this entry refer to
	Oid  common.Hash
	Name string
	Type ObjectType
	Mode common.FileMode
}
type TreeEntryCollection map[string]*TreeEntry

func newTreeEntryCollecion() TreeEntryCollection {
	tc := make(map[string]*TreeEntry)
	return tc
}

func (c TreeEntryCollection) sort() []*TreeEntry {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	entries := []*TreeEntry{}
	for _, k := range keys {
		entry := c[k]
		entries = append(entries, entry)
	}
	return entries
}

func (c TreeEntryCollection) add(te *TreeEntry) {
	c[te.Name] = te
}

func NewTreeEntry(id common.Hash, name string, mode common.FileMode) *TreeEntry {
	// var oid common.Hash
	// copy(oid[:], refId[:])
	oid := id
	var otype ObjectType

	switch mode {
	case common.Dir:
		otype = ObjectTypeTree
	case common.Regular:
		otype = ObjectTypeBlob
	default:
		otype = ObjectTypeUnknow
		panic("Not implemented.")

	}
	return &TreeEntry{
		Oid:  oid,
		Name: name,
		Type: otype,
		Mode: mode,
	}

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

func NewTree(oid common.Hash) *Tree {
	t := &Tree{
		oid: oid,
	}
	t.entries = newTreeEntryCollecion()

	return t
}

func (t *Tree) EntryHasExisted(name string) bool {
	_, ok := t.entries[name]
	return ok
}
func (t *Tree) AddEntry(entry *TreeEntry) {
	t.entries[entry.Name] = entry
}

type VisitTreeEntryFunc func(*TreeEntry)

func (t *Tree) ForEachEntry(fn VisitTreeEntryFunc) {
	keys := make([]string, 0, len(t.entries))
	for k := range t.entries {
		keys = append(keys, k)
	}

	for _, k := range keys {
		fn(t.entries[k])
	}

}

// TODO: output with format interface
func (t *Tree) ShowContent(w io.Writer) {
	entries := t.entries.sort()
	for _, entry := range entries {
		// TODO: align the output
		fmt.Fprintf(w, "%s %s %s\t%s\n", entry.Mode, entry.Type, entry.Oid.String(), entry.Name)
	}
}
