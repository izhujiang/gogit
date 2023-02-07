package plumbing

import (
	"fmt"
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
)

type CatFileOption struct {
	PrintType    bool
	PrintSize    bool
	PrintContent bool
}

func CatFile(oid common.Hash, w io.Writer, option *CatFileOption) error {
	repo := core.GetRepository()

	gObj, err := repo.Get(oid)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("obj = %+v\n", obj)

	// fmt.Fprintf(w, "object id: %s\n", obj.Id())
	if !option.PrintType && !option.PrintSize {
		option.PrintContent = true
	}

	if option.PrintType {
		fmt.Fprintf(w, "%s\n", gObj.Type())
	}
	if option.PrintSize {
		fmt.Fprintf(w, "%d\n", gObj.Size())
	}
	if option.PrintContent {
		// fmt.Fprintln(w, "object content:")
		switch gObj.Type() {
		case core.ObjectTypeBlob:
			blob := core.NewBlob(oid, "")
			blob.FromGitObject(gObj)
			blob.ShowContent(w)
		case core.ObjectTypeTree:
			tree := core.NewTree(oid, "")
			tree.FromGitObject(gObj)
			tree.ShowContent(w)
		case core.ObjectTypeCommit:
			commit := core.GitObjectToCommit(gObj)
			commit.ShowContent(w)
		default:
			panic("Not implemented")

		}
	}

	return err
}
