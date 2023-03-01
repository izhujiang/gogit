package porcelain

import (
	"io"

	"github.com/izhujiang/gogit/core"
)

// Init
func Init(w io.Writer, root string) error {
	wkspace, _ := core.GetWorkspace()

	wkspace.InitWorkspace(w, root)

	return nil
}
