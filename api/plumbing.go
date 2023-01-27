package git

import (
	"fmt"
	"io"
	"log"

	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/plumbing"
)

func HashObject(r io.Reader, w io.Writer, option *HashObjectOption) error {
	objType := core.ParseObjectType(string(option.ObjectType))
	obj, err := plumbing.HashObject(r, objType)

	if option.Write {
		repo, err := core.GetRepository()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("obj = %+v\n", obj)
		repo.Put(obj)
	}

	h := fmt.Sprintf("%s\n", obj.Id())
	w.Write([]byte(h))

	return err
}

// CatFile Provide content or type and size information for repository objects which is identified by 40 characters.
func CatFile(objectId string, w io.Writer, option *CatFileOption) error {
	// fmt.Println("CatFileOption:", *option)
	oid, err := core.NewHash(objectId)
	if err != nil {
		log.Fatal(err)
	} else {
		err := plumbing.CatFile(oid, w, (*plumbing.CatFileOption)(option))
		return err
	}

	return nil
}

func DumpObject(objectId string, w io.Writer, option *DumpObjectOption) error {
	oid, err := core.NewHash(objectId)
	if err != nil {
		log.Fatal(err)
	} else {
		err := plumbing.DumpObject(oid, w, (*plumbing.DumpObjectOption)(option))
		return err
	}

	return nil
}

// Shows one or more objects (blobs, trees, tags and commits).
func Show(name string, w io.Writer) error {
	oid, err := core.NewHash(name)
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
	oid, err := core.NewHash(treeId)
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
