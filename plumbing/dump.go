package plumbing

import (
	"io"
	"log"

	"github.com/izhujiang/gogit/core"
)

type DumpObjectOption struct {
}

func DumpObject(oid core.Hash, w io.Writer, option *DumpObjectOption) error {
	repo, err := core.GetRepository()
	if err != nil {
		log.Fatal(err)
	}

	repo.Dump(oid, w)
	return err
}
