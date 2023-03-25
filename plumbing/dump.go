package plumbing

import (
	"io"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
)

type DumpOption struct {
}

func DumpObject(w io.Writer, oid common.Hash, option *DumpOption) error {
	repo := core.GetRepository()

	repo.Dump(oid, w)
	return nil
}

func DumpIndex(w io.Writer, option *DumpOption) error {
	sa := core.GetStagingArea()
	sa.Load()
	sa.Dump(w)

	return nil
}
