package git

import (
	"fmt"
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/object"
	"github.com/izhujiang/gogit/plumbing"
)

func HashObject(r io.Reader, w io.Writer, option *HashObjectOption) error {

	pho := &plumbing.HashObjectOption{
		ObjectType: object.ParseObjectType(option.ObjectType),
		Write:      option.Write,
	}
	h, err := plumbing.HashObject(r, pho)
	hStr := fmt.Sprintf("%s\n", h)
	w.Write([]byte(hStr))

	return err
}

// CatFile Provide content or type and size information for repository objects which is identified by 40 characters.
func CatFile(objectId string, w io.Writer, option *CatFileOption) error {
	// fmt.Println("CatFileOption:", *option)
	oid, err := common.NewHash(objectId)
	if err != nil {
		log.Fatal(err)
	} else {
		err := plumbing.CatFile(oid, w, (*plumbing.CatFileOption)(option))
		return err
	}

	return nil
}

func Dump(objectId string, w io.Writer, option *DumpOption) error {
	var err error
	if objectId == "index" {
		err = plumbing.DumpIndex(w, (*plumbing.DumpOption)(option))
	} else {
		oid, err := common.NewHash(objectId)
		if err != nil {
			log.Fatal(err)
		} else {
			err = plumbing.DumpObject(oid, w, (*plumbing.DumpOption)(option))
		}
	}

	return err
}

// Shows one or more objects (blobs, trees, tags and commits).
func Show(name string, w io.Writer) error {
	oid, err := common.NewHash(name)
	if err != nil {
		log.Fatal(err)
	}
	plumbing.Show(oid, w)

	return nil
}

func Ls() error {
	return nil
}

// List the contents of a tree object
func LsTree(treeId string, w io.Writer, option *LsTreeOption) error {
	oid, err := common.NewHash(treeId)
	if err != nil {
		log.Fatal(err)
	}
	err = plumbing.LsTree(oid, w, (*plumbing.LsTreeOption)(option))
	return err
}

// MakeTree generate new tree objects from git ls-tree formatted output
func MakeTree() {

}

// LsFiles show information about files in the index and the working tree.
func LsFiles(w io.Writer, option *LsFilesOption) error {
	return plumbing.LsFiles(w, (*plumbing.LsFilesOption)(option))
}

func WriteTree(w io.Writer, option *WriteTreeOption) error {
	return plumbing.WriteTree(w, (*plumbing.WriteTreeOption)(option))
}

func ReadTree(w io.Writer, treeId string, option *ReadTreeOption) error {
	oid, err := common.NewHash(treeId)
	if err != nil {
		log.Fatal(err)
	}

	return plumbing.ReadTree(w, oid, (*plumbing.ReadTreeOption)(option))
}

func UpdateIndex(option *UpdateIndexOption) error {
	return plumbing.UpdateIndex((*plumbing.UpdateIndexOption)(option))
}
