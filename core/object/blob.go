package object

import (
	"github.com/izhujiang/gogit/common"
)

// Blob object
type Blob struct {
	oid common.Hash
	// fullpath string
	name    string
	content []byte
}

func EmptyBlob() *Blob {
	return &Blob{}
}

func NewBlob(h common.Hash, name string, content []byte) *Blob {
	b := &Blob{
		oid: h,
		// fullpath: filepath.Base(path),
		name:    name,
		content: content,
	}

	return b
}

func (b *Blob) Id() common.Hash {
	return b.oid
}

func (b *Blob) SetId(oid common.Hash) {
	b.oid = oid
}

func (b *Blob) Name() string {
	// return filepath.Base(b.fullpath)
	return b.name
}

func (b *Blob) SetName(name string) {
	// return filepath.Base(b.fullpath)
	b.name = name
}

// func (b *Blob) Path() string {
// return b.fullpath
// }

func (b *Blob) Type() ObjectType {
	return ObjectTypeBlob
}

func (b *Blob) Size() int {
	return len(b.content)

}

func (b *Blob) Mode() common.FileMode {
	return common.Regular
}

// TODO: output with format interface
func (b *Blob) Content() string {
	return string(b.content)
	// fmt.Fprintf(w, "%s", b.content)
}

// GitObject <==> Blob
func (b *Blob) FromGitObject(g *GitObject) {
	// copy(b.oid[:], g.oid[:])
	b.oid = g.Hash()
	b.content = make([]byte, g.Size())
	copy(b.content, g.Content())
}

func (b *Blob) ToGitObject() *GitObject {
	g := &GitObject{
		objectType: ObjectTypeBlob,
		size:       int64(len(b.content)),
	}

	copy(b.content, g.content)

	return g
}
