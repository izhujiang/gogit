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

	gObj, err := repo.Get(oid)
	if err != nil {
		log.Fatal(err)
	}

	if gObj.Kind() == object.Kind_Commit {

		commit := object.EmptyCommit()
		commit.FromGitObject(gObj)
		fmt.Fprintf(w, "%s", commit.Content())
	}
	return nil
}
