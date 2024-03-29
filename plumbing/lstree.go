package plumbing

import (
	"fmt"
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
func LsTree(w io.Writer, oid common.Hash, option *LsTreeOption) error {
	repo := core.GetRepository()

	gObj, err := repo.Get(oid)
	if gObj.Kind() == object.Kind_Tree {
		if option.Recurse == false {
			tree := object.GitObjectToTree(gObj)
			fmt.Fprintln(w, tree.Content())
		} else {
			panic("ls-tree recursively not implemented")
		}
	} else {
		log.Fatal("Invalid tree: ", gObj.Kind(), " ", oid)
	}

	return err

}
