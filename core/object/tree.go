package object

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
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
// // Blob and Tree Object both implemente TreeEntry interface
type TreeEntry interface {
	Object
	fs.DirEntry
}

// type DirEntry interface {
// 	Name() string
// 	IsDir() bool

// 	Type() FileMode
// 	Info() (FileInfo, error)
// }

type TreeEntryCollection []TreeEntry

func newTreeEntryCollecion() TreeEntryCollection {
	tc := make([]TreeEntry, 0)
	return tc
}

func NewTreeEntry(id common.Hash, name string, filemode common.FileMode) TreeEntry {
	// var oid common.Hash
	// copy(oid[:], refId[:])
	var te TreeEntry

	kind := FileModeToObjectKind(filemode)
	switch kind {
	case Kind_Blob:
		te = NewBlob(id, name, filemode, nil)
	case Kind_Tree:
		te = NewTree(id, name, filemode)
	default:
		fmt.Println("id: ", id, " name: ", name)
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

// Tree Object, implements Interface TreeEntry{ Object, fs.DirEntry}
type Tree struct {
	// Hash ID
	GitObject
	// fullpath string
	name string

	filemode common.FileMode

	// TODO: make sure entries ordered by name
	entries TreeEntryCollection
}

func (t *Tree) Name() string {
	return t.name
}

func (t *Tree) IsDir() bool {
	return true
}

func (t *Tree) Type() common.FileMode {
	return t.filemode
}

func (t *Tree) EntryCount() int {
	return len(t.entries)
}

// Info returns the FileInfo for the file or subdirectory described by the entry.
// The returned FileInfo may be from the time of the original directory read
// or from the time of the call to Info. If the file has been removed or renamed
// since the directory read, Info may return an error satisfying errors.Is(err, ErrNotExist).
// If the entry denotes a symbolic link, Info reports the information about the link itself,
// not the link's target.
func (t *Tree) Info() (fs.FileInfo, error) {
	info := &GitObjectInfo{}
	return info, nil
}

func (t *Tree) ZeroId() {
	t.oid = common.ZeroHash
}
func (t *Tree) SetName(name string) {
	t.name = name
}

func NewTree(oid common.Hash, name string, filemode common.FileMode) *Tree {
	return &Tree{
		GitObject: GitObject{
			oid:        oid,
			objectKind: Kind_Tree,
		},
		name:     name,
		filemode: filemode,
		entries:  newTreeEntryCollecion(),
	}
}
func EmptyTree() *Tree {
	return &Tree{
		entries: newTreeEntryCollecion(),
	}
}

type WalkTreeEntryFunc func(TreeEntry) error

func (t *Tree) ForEach(fn WalkTreeEntryFunc) {
	for _, e := range t.entries {
		err := fn(e)

		if err == filepath.SkipDir {
			break
		}
	}
}

func (t *Tree) Subtree(subtreeName string) *Tree {
	for _, e := range t.entries {
		if e.Type() == common.Dir && e.Name() == subtreeName {
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

	// Invalidate oid when the entries are changed
	t.oid = common.ZeroHash
}

func (t *Tree) RemoveEmptyEntries() {
	es := make([]TreeEntry, 0, t.EntryCount())
	for _, e := range t.entries {
		switch e.Kind() {
		case Kind_Blob:
			es = append(es, e)
		case Kind_Tree:
			subT := e.(*Tree)
			if subT.EntryCount() != 0 {
				es = append(es, e)
			}
		}
	}
	t.entries = es

	// Invalidate oid when the entries are changed
	t.oid = common.ZeroHash
}

func (t *Tree) Sort() {
	// sort entries
	entries := t.entries
	// if order == order_Ascending {
	sort.SliceStable(entries, func(i, j int) bool {
		return strings.Compare(entries[i].Name(), entries[j].Name()) < 0
	})
}
func (t *Tree) Hash() common.Hash {
	t.composeContent()
	return t.GitObject.Hash()
}

func (t *Tree) Content() string {
	buf := &bytes.Buffer{}
	for _, e := range t.entries {
		// TODO: align the output
		// o := e.(Object)
		fmt.Fprintf(buf, "%s %s %s\t%s\n", common.FileModeToString(e.Type()), e.Kind(), e.Id(), e.Name())
	}

	return string(buf.Bytes())
}

// GitObject ==> Tree,fitll Tree using GotObject from repository
func (t *Tree) FromGitObject(g *GitObject) {
	t.GitObject = *g
	t.parseContent()
}

func GitObjectToTree(g *GitObject) *Tree {
	t := &Tree{
		GitObject: *g,
	}
	t.parseContent()

	return t
}

func (t *Tree) Serialize(w io.Writer) error {
	t.composeContent()

	return t.GitObject.Serialize(w)
}

func (t *Tree) Deserialize(r io.Reader) error {
	err := t.GitObject.Deserialize(r)

	if err != nil {
		return err
	}

	t.parseContent()
	return nil
}

func (t *Tree) parseContent() {
	buf := bytes.NewBuffer(t.content)
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

	t.entries = entries
}

func (t *Tree) composeContent() {
	buf := &bytes.Buffer{}

	entries := t.entries
	for _, e := range entries {
		mode := strings.TrimLeft(common.FileModeToString(e.Type()), "0 ")
		buf.WriteString(mode)
		buf.WriteByte(common.SPACE)
		buf.WriteString(e.Name())
		buf.WriteByte(common.NUL)
		id := e.Id()
		buf.Write(id[:])
	}
	t.content = buf.Bytes()
}
