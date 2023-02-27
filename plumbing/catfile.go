package plumbing

import (
	"fmt"
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
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
		switch gObj.Type() {
		case object.ObjectTypeBlob:
			blob := object.GitObjectToBlob(gObj)
			blob.ShowContent(w)
		case object.ObjectTypeTree:
			tree := object.GitObjectToTree(gObj)
			tree.ShowContent(w)
		case object.ObjectTypeCommit:
			commit := object.GitObjectToCommit(gObj)
			commit.ShowContent(w)
		default:
			panic("Not implemented")

		}
	}

	return err
}
