package git

import (
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/porcelain"
)

// Local  Operations

// Init
func Init(w io.Writer, root string) error {
	return porcelain.Init(w, root)
}

// Add file contents to the index
// This command updates the index using the current content found in the working tree, to prepare the content staged for the next commit. It
// typically adds the current content of existing paths as a whole, but with some options it can also be used to add content with only part of
// the changes made to the working tree files applied, or remove paths that do not exist in the working tree anymore.
//
// The "index" holds a snapshot of the content of the working tree, and it is this snapshot that is taken as the contents of the next commit.
// Thus after making any changes to the working tree, and before running the commit command, you must use the add command to add any new or
// modified files to the index.

func Add(paths []string, option *AddOption) error {
	return porcelain.Add(paths)
}

func Remove(paths []string, option *RemoveOption) error {
	return porcelain.Remove(paths, option)
}

func Reset() error {
	return nil
}

func Commit(w io.Writer, option *CommitOption) error {
	return porcelain.Commit(w, (*porcelain.CommitOption)(option))
}

func Status() error {
	return nil
}

func Config() error {
	return nil
}

func Branch() error {
	return nil
}

func Checkout() error {
	return nil
}

func Merge() error {
	return nil
}

func Stash() error {
	return nil
}

func Log(w io.Writer, commitId string, option *LogOption) error {
	oid, err := common.NewHash(commitId)
	if err != nil {
		log.Fatal(err)
	}
	return porcelain.Log(w, oid, (*porcelain.LogOption)(option))
}
