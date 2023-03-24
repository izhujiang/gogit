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

	g, err := repo.Get(oid)
	if err != nil {
		log.Fatal(err)
	}

	if !option.PrintType && !option.PrintSize {
		option.PrintContent = true
	}

	if option.PrintType {
		fmt.Fprintf(w, "%s\n", g.Kind())
	}
	if option.PrintSize {
		fmt.Fprintf(w, "%d\n", g.Size())
	}
	if option.PrintContent {
		switch g.Kind() {
		case object.Kind_Blob:
			blob := object.GitObjectToBlob(g)
			fmt.Fprintf(w, "%s", blob.Content())
		case object.Kind_Tree:
			tree := object.GitObjectToTree(g)
			fmt.Fprintf(w, "%s", tree.Content())
		case object.Kind_Commit:
			commit := object.GitObjectToCommit(g)
			fmt.Fprintf(w, "%s", commit.Content())
		default:
			panic("Not implemented")

		}
	}

	return err
}
