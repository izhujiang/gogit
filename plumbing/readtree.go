package plumbing

import (
	"io"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
)

type ReadTreeOption struct {
	Prefix string
}

// Reads tree information into the index.
func ReadTree(w io.Writer, oid common.Hash, option *ReadTreeOption) error {
	sa := core.GetStagingArea()
	if option.Prefix == "" {
		sa.ReadTree(oid, "")
	} else {
		// TODO: add an entry to the tree, and save it to stage area
		panic("Not implemented")
	}

	return nil
}
