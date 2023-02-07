package core

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/internal/filemode"
)

// 100644 blob ec1871edcbfdc0d17ef498030e7ca676f291393d	LICENSE
type TreeEntry struct {
	Mode filemode.FileMode
	Type ObjectType

	// Id of subtree or blob,  which this entry refer to
	Oid  common.Hash
	Name string
}
type TreeEntryCollection map[string]*TreeEntry

func newTreeEntryCollecion() TreeEntryCollection {
	tc := make(map[string]*TreeEntry)
	return tc
}

func (tc TreeEntryCollection) sort() []*TreeEntry {
	keys := make([]string, 0, len(tc))
	for k := range tc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	entries := []*TreeEntry{}
	for _, k := range keys {
		entry := tc[k]
		entries = append(entries, entry)
	}
	return entries
}

func (tc TreeEntryCollection) add(name string, te *TreeEntry) {
	tc[name] = te
}

func newTreeEntry(mode filemode.FileMode, name string, refId common.Hash) *TreeEntry {
	var oid common.Hash
	copy(oid[:], refId[:])
	var otype ObjectType

	switch mode {
	case filemode.Dir:
		otype = ObjectTypeTree
	case filemode.Regular:
		otype = ObjectTypeBlob
	default:
		otype = ObjectTypeUnknow
		panic("Not implemented.")

	}
	return &TreeEntry{
		Mode: mode,
		Type: otype,
		Name: name,
		Oid:  oid,
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
	// optional: full path of the Tree
	name string

	// TODO: make sure entries ordered by name
	// entries map[string]*TreeEntry
	entries TreeEntryCollection
}

func (t *Tree) Id() common.Hash {
	var h common.Hash
	copy(h[:], t.oid[:])
	return h
}

func (t *Tree) Name() string {
	return t.name
}

func NewTree(oid common.Hash, name string) *Tree {
	t := &Tree{}
	copy(t.oid[:], oid[:])
	t.name = name
	t.entries = newTreeEntryCollecion()

	return t
}

func (t *Tree) EntryHasExisted(name string) bool {
	_, ok := t.entries[name]
	return ok
}

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

		fm, _ := filemode.New(mode)
		entry := newTreeEntry(fm, fileName, oid)
		entries.add(fileName, entry)
	}

	t.entries = entries
}

func (t *Tree) ToGitObject() *GitObject {
	w := &bytes.Buffer{}

	entries := t.entries.sort()
	for _, entry := range entries {
		mode := strings.TrimLeft(entry.Mode.String(), "0 ")
		w.WriteString(mode)
		w.WriteByte(0x20)
		w.WriteString(entry.Name)
		w.WriteByte(0x00)
		w.Write(entry.Oid[:])
	}

	content := w.Bytes()

	g := &GitObject{
		objectType: ObjectTypeTree,
		size:       int64(len(content)),
		content:    content,
	}
	return g
}

// TODO: output with format interface
func (t *Tree) ShowContent(w io.Writer) {
	entries := t.entries.sort()
	for _, entry := range entries {
		// TODO: align the output
		fmt.Fprintf(w, "%s %s %s\t%s\n", entry.Mode, entry.Type, entry.Oid.String(), entry.Name)
	}
}
