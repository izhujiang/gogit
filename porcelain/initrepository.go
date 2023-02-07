package porcelain

import (
	"github.com/izhujiang/gogit/core"
)

// Init
func Init(root string) error {
	wkspace, _ := core.GetWorkspace()

	wkspace.InitWorkspace(root)

	return nil
}
