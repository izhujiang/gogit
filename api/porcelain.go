package git

import (
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/porcelain"
)

// Local  Operations

// Init
func Init(root string) error {
	return porcelain.Init(root)
}

// Add file contents to the index
// This command updates the index using the current content found in the working tree, to prepare the content staged for the next commit. It
// typically adds the current content of existing paths as a whole, but with some options it can also be used to add content with only part of
// the changes made to the working tree files applied, or remove paths that do not exist in the working tree anymore.
//
// The "index" holds a snapshot of the content of the working tree, and it is this snapshot that is taken as the contents of the next commit.
// Thus after making any changes to the working tree, and before running the commit command, you must use the add command to add any new or
// modified files to the index.
func Add(path string, option *AddOption) (string, error) {
	oid, err := porcelain.Add(path)
	if err != nil {
		log.Fatal(err)
		return common.InvalidObjectId.String(), nil
	}

	return oid.String(), err
}

func Reset() error {
	return nil
}

func Commit() error {
	return nil
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

func Log() error {
	return nil
}

func Rm() error {
	return nil
}
