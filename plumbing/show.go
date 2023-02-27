package plumbing

import (
	"io"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

// Shows one or more objects (blobs, trees, tags and commits).
// For commits it shows the log message and textual diff. It also presents the merge commit in a special format as produced by git diff-tree --cc.
// For tags, it shows the tag message and the referenced objects.
// For trees, it shows the names (equivalent to git ls-tree with --name-only).
// For plain blobs, it shows the plain contents.
// The command takes options applicable to the git diff-tree command to control how the changes the commit introduces are shown.
func Show(oid common.Hash, w io.Writer) error {
	// TODO: need to be refactored
	repo := core.GetRepository()

	gObj, err := repo.Get(oid)
	switch gObj.Type() {
	case object.ObjectTypeBlob:
		blob := object.GitObjectToBlob(gObj)
		blob.ShowContent(w)
	case object.ObjectTypeTree:
		tree := object.GitObjectToTree(gObj)
		tree.ShowContent(w)
	case object.ObjectTypeCommit:
		commit := object.GitObjectToCommit(gObj)
		commit.ShowContent(w)
	case object.ObjectTypeTag:
		panic("Not implemented")
	default:
		panic("Not implemented")
	}

	return err
}
