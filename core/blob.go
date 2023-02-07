package core

import (
	"fmt"
	"io"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/internal/filemode"
)

// Blob object
type Blob struct {
	// gObj *GitObject
	oid     common.Hash
	name    string
	content []byte
}

func NewBlob(h common.Hash, name string) *Blob {
	b := &Blob{
		name: name,
	}
	copy(b.oid[:], h[:])

	return b
}

func (b *Blob) Id() common.Hash {
	return b.oid
}

func (b *Blob) Type() ObjectType {
	return ObjectTypeBlob
}

func (b *Blob) Name() string {
	return b.name
}

func (b *Blob) Size() int {
	return len(b.content)

}

func (b *Blob) Mode() filemode.FileMode {
	panic("Not implemented")
}

func (b *Blob) Entries() []TreeEntry {
	return []TreeEntry{}
}

func (b *Blob) FromGitObject(g *GitObject) {
	// copy(b.oid[:], g.oid[:])
	copy(b.content, g.Content())
}

func (b *Blob) ToGitObject() *GitObject {
	g := &GitObject{
		objectType: ObjectTypeBlob,
		size:       int64(len(b.content)),
	}

	// copy(b.oid[:], g.oid[:])
	copy(b.content, g.content)

	return g
}

// TODO: output with format interface
func (b *Blob) ShowContent(w io.Writer) {
	// fmt.Fprintf(w, "%s", b.gObj.Content())
	fmt.Fprintf(w, "%s", b.content)
}
