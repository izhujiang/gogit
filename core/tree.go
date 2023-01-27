package core

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/izhujiang/gogit/core/internal/filemode"
)

type TreeEntry struct {
	Mode          filemode.FileMode
	Type          ObjectType
	Name          string
	FileOrSubtree Hash
}

func newTreeEntry(mode string, name string, ftId Hash) *TreeEntry {
	m, _ := filemode.New(mode)
	var otype ObjectType

	switch m {
	case filemode.Dir:
		otype = ObjectTypeTree
	case filemode.Regular:
		otype = ObjectTypeBlob
	default:
		otype = ObjectTypeUnknow
		panic("Not implemented.")

	}

	return &TreeEntry{
		Mode:          m,
		Type:          otype,
		Name:          name,
		FileOrSubtree: ftId,
	}

}

type Tree struct {
	gObj    *GitObject
	entries []*TreeEntry
	// content
	// mode(string)0x20name(string)0x00HASH[20]
}

func NewTree(g *GitObject) *Tree {
	tree := &Tree{
		gObj: g,
	}

	tree.parseContent(g.content)

	return tree
}

func (t *Tree) parseContent(content []byte) {
	r := bytes.NewBuffer(content)
	entries := []*TreeEntry{}

	for {
		mode, err := r.ReadString(0x20)
		mode = strings.Trim(mode, " ")
		if err == io.EOF {
			break
		}
		name, _ := r.ReadString(0x00)
		var oid Hash
		_, _ = r.Read(oid[:])

		// fmt.Println("content: ", content)
		// fmt.Fscanf(r, "%s\x20%s\x00", &mode, &name)
		// fmt.Println(name, mode, oid)

		entry := newTreeEntry(mode, name, oid)
		entries = append(entries, entry)
	}

	t.entries = entries
}

// TODO: output with format interface
func (t *Tree) ShowContent(w io.Writer) {
	// buf := &bytes.Buffer{}
	for _, v := range t.entries {
		// TODO: align the output
		// fmt.Fprintf(buf, "%s %s %s\t%s\n", v.Mode, v.Type, v.FileOrSubtree.String(), v.Name)
		fmt.Fprintf(w, "%s %s %s\t%s\n", v.Mode, v.Type, v.FileOrSubtree.String(), v.Name)
	}

	// return buf.Bytes()
}
