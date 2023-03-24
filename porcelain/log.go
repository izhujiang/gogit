package porcelain

import (
	"fmt"
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

type LogOption struct {
	Stat bool
}

func Log(w io.Writer, oid common.Hash, option *LogOption) error {
	repo := core.GetRepository()

	g, err := repo.Get(oid)
	if err != nil {
		log.Fatal(err)
	}

	if g.Kind() == object.Kind_Commit {
		commit := object.GitObjectToCommit(g)
		fmt.Fprintf(w, "%s", commit.Content())
	}
	return nil
}
