package plumbing

import (
	"io"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
)

type ReadTreeOption struct {
	HasPrefix bool
	Prefix    string
}

// Reads tree information into the index.
func ReadTree(w io.Writer, oid common.Hash, option *ReadTreeOption) error {
	sa := core.GetStagingArea()

	var eraseOriginal bool
	if option.HasPrefix == false {
		eraseOriginal = true

	}
	return sa.ReadTree(oid, option.Prefix, eraseOriginal)
}
