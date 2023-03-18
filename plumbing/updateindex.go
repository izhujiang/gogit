package plumbing

import (
	"log"
	"os"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

type UpdateIndexOption struct {
	Op   string
	Path string
	Args map[string]string // Args["oid"], Args["mode"], Args["file"] is valid only Path is ""
}

// Register file contents in the working tree to the index
func UpdateIndex(option *UpdateIndexOption) error {
	sa := core.GetStagingArea()
	switch option.Op {
	case "replace", "add":
		if option.Path != "" {
			path := option.Path
			f, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			hoOption := &HashObjectOption{
				ObjectType: object.Kind_Blob,
				Write:      true,
			}
			oid, err := HashObject(f, hoOption)

			sa.UpdateIndex(oid, path)
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

	return nil
}
