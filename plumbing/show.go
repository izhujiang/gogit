package plumbing

import (
	"io"
	"log"

	"github.com/izhujiang/gogit/core"
)

// Shows one or more objects (blobs, trees, tags and commits).
// For commits it shows the log message and textual diff. It also presents the merge commit in a special format as produced by git diff-tree --cc.
// For tags, it shows the tag message and the referenced objects.
// For trees, it shows the names (equivalent to git ls-tree with --name-only).
// For plain blobs, it shows the plain contents.
// The command takes options applicable to the git diff-tree command to control how the changes the commit introduces are shown.
func Show(oid core.Hash, w io.Writer) error {
	// TODO: need to be refactored
	repo, err := core.GetRepository()
	if err != nil {
		log.Fatal(err)
	}

	gObj, err := repo.Get(oid)
	switch gObj.Type() {
	case core.ObjectTypeBlob:
		blob := core.NewBlob(gObj)
		blob.ShowContent(w)
	case core.ObjectTypeTree:
		tree := core.NewTree(gObj)
		tree.ShowContent(w)
	case core.ObjectTypeCommit:
		commit := core.NewCommit(gObj)
		commit.ShowContent(w)
	case core.ObjectTypeTag:
		panic("Not implemented")
	default:
		panic("Not implemented")
	}

	return nil
}
