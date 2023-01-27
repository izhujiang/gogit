package plumbing

import (
	"io"
	"log"

	"github.com/izhujiang/gogit/core"
)

type LsTreeOption struct {
	Recurse bool
}

// LsTree list the contents of a tree object.
// Lists the contents of a given tree object, like what "/bin/ls -a" does in the current working directory.
func LsTree(oid core.Hash, w io.Writer, option *LsTreeOption) error {
	repo, err := core.GetRepository()
	if err != nil {
		log.Fatal(err)
	}
	gObj, err := repo.Get(oid)
	if gObj.Type() == core.ObjectTypeTree {
		if option.Recurse == false {
			tree := core.NewTree(gObj)
			tree.ShowContent(w)
		} else {
			panic("ls-tree recursively not implemented")
		}
	} else {
		log.Fatal("Invalid tree: ", gObj.Type().String(), " ", gObj.Id().String())
	}

	return err

}
