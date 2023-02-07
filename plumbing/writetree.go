package plumbing

import (
	"fmt"
	"io"
	"log"

	"github.com/izhujiang/gogit/core"
)

type WriteTreeOption struct {
	prefix string
}

// WriteTree create a tree object from the current index
func WriteTree(w io.Writer, option *WriteTreeOption) error {
	sa := core.GetStagingArea()
	tid, err := sa.WriteTree()

	if err != nil {
		log.Fatal(err)
	}

	// fmt.Fprintln(w, "Tree Id:")
	fmt.Fprintf(w, "\t%s\n", tid)
	return err
}
