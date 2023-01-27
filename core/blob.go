package core

import (
	"fmt"
	"io"
)

// Blob object
type Blob struct {
	gObj *GitObject
}

func NewBlob(g *GitObject) *Blob {
	return &Blob{
		gObj: g,
	}
}

// TODO: output with format interface
func (b *Blob) ShowContent(w io.Writer) {
	fmt.Fprintf(w, "%s", b.gObj.Content())
}
