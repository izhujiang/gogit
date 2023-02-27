package plumbing

import (
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

type LsTreeOption struct {
	Recurse bool
}

// LsTree list the contents of a tree object.
// Lists the contents of a given tree object, like what "/bin/ls -a" does in the current working directory.
func LsTree(oid common.Hash, w io.Writer, option *LsTreeOption) error {
	repo := core.GetRepository()

	gObj, err := repo.Get(oid)
	if gObj.Type() == object.ObjectTypeTree {
		if option.Recurse == false {
			tree := object.GitObjectToTree(gObj)
			tree.ShowContent(w)
		} else {
			panic("ls-tree recursively not implemented")
		}
	} else {
		log.Fatal("Invalid tree: ", gObj.Type().String(), " ", oid.String())
	}

	return err

}
