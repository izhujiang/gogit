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

func CatFile(w io.Writer, oid common.Hash, option *CatFileOption) error {
	repo := core.GetRepository()

	gObj, err := repo.Get(oid)
	if err != nil {
		log.Fatal(err)
	}

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
			blob := object.EmptyBlob()
			blob.FromGitObject(gObj)
			fmt.Fprintf(w, "%s", blob.Content())
		case object.ObjectTypeTree:
			tree := object.EmptyTree()
			tree.FromGitObject(gObj)
			fmt.Fprintf(w, "%s", tree.Content())
		case object.ObjectTypeCommit:
			commit := object.EmptyCommit()
			commit.FromGitObject(gObj)
			fmt.Fprintf(w, "%s", commit.Content())
		default:
			panic("Not implemented")

		}
	}

	return err
}
