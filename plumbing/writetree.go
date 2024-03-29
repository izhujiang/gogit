package plumbing

import (
	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
)

type WriteTreeOption struct {
	// prefix string
}

// WriteTree create a tree object from the current index (.git/index file)
func WriteTree(option *WriteTreeOption) (common.Hash, error) {
	sa := core.GetStagingArea()
	sa.Load()
	tid, err := sa.WriteTree()
	if err != nil {
		return tid, err
	}
	err = sa.Save()
	return tid, err
}
