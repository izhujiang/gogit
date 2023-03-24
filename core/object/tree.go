package object

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/izhujiang/gogit/common"
)

const (
	default_tree_entry_capacity = 64
)

// 100644 blob ec1871edcbfdc0d17ef498030e7ca676f291393d	LICENSE
// // Blob and Tree Object both implemente TreeEntry interface
type TreeEntry struct {
	Oid      common.Hash
	Kind     ObjectKind
	Name     string
	Filemode common.FileMode

	// Pointer to Subtree(*Tree) or Blob(*Blob) identified by oid and implement common Object interface
	Pointer Object
}

func NewTreeEntry(id common.Hash, name string, filemode common.FileMode) *TreeEntry {
	kind := ObjectKindFromFilemode(filemode)

	e := &TreeEntry{
		Oid:      id,
		Kind:     kind,
		Name:     name,
		Filemode: filemode,
	}

	return e
}

type TreeEntryCollection []*TreeEntry

func newTreeEntryCollecion() TreeEntryCollection {
	tc := make([]*TreeEntry, 0, default_tree_entry_capacity)
	return tc
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

// Tree implements Object interface
type Tree struct {
	// Hash ID
	oid common.Hash

	// TODO: make sure entries ordered by name
	entries TreeEntryCollection
}

func NewTree(oid common.Hash) *Tree {
	return &Tree{
		oid:     oid,
		entries: newTreeEntryCollecion(),
	}
}

func EmptyTree() *Tree {
	return &Tree{
		entries: newTreeEntryCollecion(),
	}
}

func (t *Tree) Id() common.Hash {
	return t.oid
}

func (t *Tree) Kind() ObjectKind {
	return Kind_Tree
}

func (t *Tree) EntryCount() int {
	return len(t.entries)
}

func (t *Tree) ZeroId() {
	t.oid = common.ZeroHash
}

type WalkTreeEntryFunc func(*TreeEntry) error

func (t *Tree) ForEach(fn WalkTreeEntryFunc) {
	for _, e := range t.entries {
		err := fn(e)

		if err == filepath.SkipDir {
			break
		}
	}
}

func (t *Tree) Find(name string) *TreeEntry {
	for _, e := range t.entries {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func (t *Tree) Subtree(name string) *TreeEntry {
	for _, e := range t.entries {
		if e.Kind == Kind_Tree && e.Name == name {
			return e
		}
	}

	return nil
}

func (t *Tree) Append(entry *TreeEntry) {
	t.entries = append(t.entries, entry)

	// do invalidate oid manually
	// t.oid = common.ZeroHash
}

func (t *Tree) Sort() {
	entries := t.entries
	sort.SliceStable(entries, func(i, j int) bool {
		return strings.Compare(entries[i].Name, entries[j].Name) < 0
	})

	// do invalidate oid manually
	// t.oid = common.ZeroHash
}

// Update, remove extra empty subtrees and update tree entries which are async with subtree
// Caution: empty tree entries will be remove and
func (t *Tree) UpdateEntryState() {
	entrychanged := false

	es := make([]*TreeEntry, 0, t.EntryCount())
	for _, e := range t.entries {
		switch e.Kind {
		case Kind_Blob:
			es = append(es, e)

		case Kind_Tree:
			if e.Pointer == nil { // do nothing, leave it alone
				es = append(es, e)
			} else {
				subT := e.Pointer.(*Tree)
				if subT.EntryCount() != 0 {
					if e.Oid != subT.Id() {
						e.Oid = subT.Id()
						entrychanged = true
					}
					es = append(es, e)
				} else { // filter out the empty entry
					entrychanged = true
				}
			}

		} // endof switch
	}
	t.entries = es

	if entrychanged {
		t.oid = common.ZeroHash
	}
}

func (t *Tree) Hash() common.Hash {
	c := t.contentToBytes()
	t.oid = common.HashObject(t.Kind().String(), c)
	return t.oid
}

func (t *Tree) Content() string {
	buf := &bytes.Buffer{}
	for _, e := range t.entries {
		// TODO: align the output
		// o := e.(Object)
		fmt.Fprintf(buf, "%s %s %s\t%s\n", common.FileModeToString(e.Filemode), e.Kind, e.Oid, e.Name)
	}

	return string(buf.Bytes())
}

// GitObject ==> Tree,fitll Tree using GotObject from repository
func GitObjectToTree(g *GitObject) *Tree {
	buf := bytes.NewBuffer(g.content)
	entries := newTreeEntryCollecion()

	for {
		mode, err := buf.ReadString(common.SPACE)
		mode = strings.Trim(mode, " ")
		if err == io.EOF {
			break
		}
		name, _ := buf.ReadBytes(common.NUL)
		fileName := string(name[:len(name)-1])
		var oid common.Hash
		_, _ = buf.Read(oid[:])

		fm, _ := common.NewFileMode(mode)
		// entry := NewTreeEntry(oid, filepath.Join(t.fullpath, fileName), fm)
		entry := NewTreeEntry(oid, fileName, fm)

		entries = append(entries, entry)
	}

	t := &Tree{
		oid:     g.oid,
		entries: entries,
	}

	return t
}

func (t *Tree) ToGitObject() *GitObject {
	g := NewGitObject(Kind_Tree, t.contentToBytes())

	// TODO: what to do if t.oid !!= g.oid, in case t.oid == ZeroHash
	if t.oid != g.oid {
		log.Fatal(ErrInvalidObject, t.oid, g.oid)
	}

	return g
}

func (t *Tree) contentToBytes() []byte {
	buf := &bytes.Buffer{}

	for _, e := range t.entries {
		mode := strings.TrimLeft(common.FileModeToString(e.Filemode), "0 ")
		buf.WriteString(mode)
		buf.WriteByte(common.SPACE)
		buf.WriteString(e.Name)
		buf.WriteByte(common.NUL)
		buf.Write(e.Oid[:])
	}

	return buf.Bytes()
}
