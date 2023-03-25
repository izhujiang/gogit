package plumbing

import (
	"os"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
)

type UpdateIndexOption struct {
	Op   string
	Path string
	Args map[string]string // Args["oid"], Args["mode"], Args["file"] is valid only Path is ""
}

// Register file contents in the working tree to the index
func UpdateIndex(option *UpdateIndexOption) error {
	sa := core.GetStagingArea()
	sa.Load()

	switch option.Op {
	case "replace", "add":
		if option.Path != "" {
			path := option.Path
			sa.UpdateIndex(path)
		} else {
			oid, err := common.NewHash(option.Args["oid"])
			if err != nil {
				return err
			}
			mode, err := common.NewFileMode(option.Args["mode"])
			if err != nil {
				return err
			}
			path := option.Args["file"]
			// TODO: check if oid is valid
			sa.UpdateIndexFromCache(oid, path, mode)
		}

	case "remove":
		_, err := os.Stat(option.Path)
		if err == nil { // file exist
			return nil
		}
		sa.UpdateIndexRemove(option.Path)

	default:
		msg := "updateIndex with " + option.Op + " not implemented."
		panic(msg)
	}

	err := sa.Save()

	return err
}
