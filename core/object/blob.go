package object

import (
	"fmt"
	"io"

	"github.com/izhujiang/gogit/common"
)

// Blob object
type Blob struct {
	oid     common.Hash
	content []byte
}

func NewBlob(h common.Hash, content []byte) *Blob {
	b := &Blob{
		oid:     h,
		content: content,
	}

	return b
}

func (b *Blob) Id() common.Hash {
	return b.oid
}

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
func (b *Blob) ShowContent(w io.Writer) {
	// fmt.Fprintf(w, "%s", b.gObj.Content())
	fmt.Fprintf(w, "%s", b.content)
}
