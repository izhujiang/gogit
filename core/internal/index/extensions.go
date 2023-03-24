package index

import (
	"github.com/izhujiang/gogit/common"
)

type ResolveUndoEntry struct {
	Path   string
	Stages map[Stage]common.Hash
}

type ResolveUndo struct {
	Entries []*ResolveUndoEntry
}

func newResolveUndo() *ResolveUndo {
	return &ResolveUndo{Entries: make([]*ResolveUndoEntry, 0)}

}

// unknown extention
type Extension struct {
	Signature []byte // If the first byte is 'A'..'Z' the extension is optional and can be ignored.
	Size      uint32 // 32-bit size of the extension
	Data      []byte // Extension data
}
