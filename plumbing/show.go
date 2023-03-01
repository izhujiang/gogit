package plumbing

import (
	"fmt"
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
func Show(w io.Writer, oid common.Hash) error {
	// TODO: need to be refactored
	repo := core.GetRepository()

	gObj, err := repo.Get(oid)
	var entity object.GitEntity
	switch gObj.Type() {
	case object.ObjectTypeBlob:
		entity = object.EmptyBlob()
	case object.ObjectTypeTree:
		entity = object.EmptyTree()
	case object.ObjectTypeCommit:
		entity = object.EmptyCommit()
	case object.ObjectTypeTag:
		panic("Not implemented")
	default:
		panic("Not implemented")
	}
	entity.FromGitObject(gObj)
	fmt.Fprintln(w, entity.Content())

	return err
}
