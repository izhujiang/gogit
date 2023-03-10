package git

import (
	"fmt"
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/object"
	"github.com/izhujiang/gogit/plumbing"
)

func HashObject(w io.Writer, r io.Reader, option *HashObjectOption) error {
	ho := &plumbing.HashObjectOption{
		ObjectType: object.ParseObjectType(option.ObjectType),
		Write:      option.Write,
	}
	h, err := plumbing.HashObject(r, ho)
	hStr := fmt.Sprintf("%s\n", h)
	w.Write([]byte(hStr))

	return err
}

// CatFile Provide content or type and size information for repository objects which is identified by 40 characters.
func CatFile(w io.Writer, objectId string, option *CatFileOption) error {
	// fmt.Println("CatFileOption:", *option)
	oid, err := common.NewHash(objectId)
	if err != nil {
		log.Fatal(err)
	} else {
		err := plumbing.CatFile(w, oid, (*plumbing.CatFileOption)(option))
		return err
	}

	return nil
}

func Dump(w io.Writer, objectId string, option *DumpOption) error {
	var err error
	if objectId == "index" {
		err = plumbing.DumpIndex(w, (*plumbing.DumpOption)(option))
	} else {
		oid, err := common.NewHash(objectId)
		if err != nil {
			log.Fatal(err)
		} else {
			err = plumbing.DumpObject(w, oid, (*plumbing.DumpOption)(option))
		}
	}

	return err
}

// Shows one or more objects (blobs, trees, tags and commits).
func Show(w io.Writer, name string) error {
	oid, err := common.NewHash(name)
	if err != nil {
		log.Fatal(err)
	}
	plumbing.Show(w, oid)

	return nil
}

func Ls() error {
	return nil
}

// List the contents of a tree object
func LsTree(w io.Writer, treeId string, option *LsTreeOption) error {
	oid, err := common.NewHash(treeId)
	if err != nil {
		log.Fatal(err)
	}
	err = plumbing.LsTree(w, oid, (*plumbing.LsTreeOption)(option))
	return err
}

// MakeTree generate new tree objects from git ls-tree formatted output
func MakeTree() {

}

// LsFiles show information about files in the index and the working tree.
func LsFiles(w io.Writer, option *LsFilesOption) error {
	return plumbing.LsFiles(w, (*plumbing.LsFilesOption)(option))
}

func UpdateIndex(option *UpdateIndexOption) error {
	return plumbing.UpdateIndex((*plumbing.UpdateIndexOption)(option))
}

func WriteTree(w io.Writer, option *WriteTreeOption) error {
	tid, err := plumbing.WriteTree((*plumbing.WriteTreeOption)(option))
	fmt.Fprintf(w, "%s\n", tid)
	return err
}

func ReadTree(w io.Writer, treeId string, option *ReadTreeOption) error {
	oid, err := common.NewHash(treeId)
	if err != nil {
		log.Fatal(err)
	}

	return plumbing.ReadTree(w, oid, (*plumbing.ReadTreeOption)(option))
}

func CommitTree(w io.Writer, treeId string, option *CommitTreeOption) error {
	oid, err := common.NewHash(treeId)
	if err != nil {
		log.Fatal(err)
	}

	parents := make([]common.Hash, 0, len(option.Parents))
	for _, p := range option.Parents {
		pid, err := common.NewHash(p)
		if err == nil {
			parents = append(parents, pid)
		}
	}

	cto := &plumbing.CommitTreeOption{
		Parents: parents,
		Message: option.Message,
	}

	commitId, err := plumbing.CommitTree(oid, cto)
	w.Write([]byte(commitId.String()))

	return err
}
