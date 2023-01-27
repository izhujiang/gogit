package plumbing

import (
	"fmt"
	"io"
	"log"

	"github.com/izhujiang/gogit/core"
)

type CatFileOption struct {
	PrintType    bool
	PrintSize    bool
	PrintContent bool
}

func CatFile(oid core.Hash, w io.Writer, option *CatFileOption) error {
	repo, err := core.GetRepository()
	if err != nil {
		log.Fatal(err)
	}

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
			blob := core.NewBlob(gObj)
			blob.ShowContent(w)
		case core.ObjectTypeTree:
			tree := core.NewTree(gObj)
			tree.ShowContent(w)
		case core.ObjectTypeCommit:
			commit := core.NewCommit(gObj)
			commit.ShowContent(w)
		default:
			panic("Not implemented")

		}
	}

	return err
}
