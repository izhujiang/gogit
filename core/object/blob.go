package object

import (
	"github.com/izhujiang/gogit/common"
)

// Blob object, implements Interface TreeEntry{ Object, fs.DirEntry}
type Blob struct {
	oid     common.Hash
	content []byte
}

func EmptyBlob() *Blob {
	return &Blob{}
}

func NewBlob(oid common.Hash, content []byte) *Blob {
	c := make([]byte, len(content))
	copy(c, content)
	b := &Blob{
		oid:     oid,
		content: c,
	}

	return b
}

func (b *Blob) Id() common.Hash {
	return b.oid
}

func (b *Blob) Kind() ObjectKind {
	return Kind_Blob
}

func (b *Blob) Content() string {
	return string(b.content)
}

func (b *Blob) Hash() common.Hash {
	b.oid = common.HashObject(b.Kind().String(), b.content)
	return b.oid
}

func (b *Blob) FromGitObject(g *GitObject) {
	b.oid = g.oid

	c := make([]byte, len(g.content))
	copy(b.content, c)
}

func GitObjectToBlob(g *GitObject) *Blob {
	c := make([]byte, len(g.content))
	copy(c, g.content)
	b := &Blob{
		oid:     g.oid,
		content: c,
	}

	return b
}
