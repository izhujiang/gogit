package plumbing

import (
	"fmt"
	"time"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

type CommitTreeOption struct {
	// the id of a parent commit object
	Parents []common.Hash
	// A paragraph in the commit log message.
	Message string
}

// Reads tree information into the index.
func CommitTree(tree common.Hash, option *CommitTreeOption) (common.Hash, error) {
	t := time.Now()
	u := t.Unix()
	if u < 0 {
		u = 0
	}
	s := fmt.Sprintf("%d %s", u, t.Format("-0700"))
	email := "Jiang Zhu <m.zhujiang@gmail.com>"

	// TODO: using config info
	author := fmt.Sprintf("%s %s", email, s)
	committer := fmt.Sprintf("%s %s", email, s)
	parents := option.Parents

	c := object.NewCommit(
		common.ZeroHash,
		tree,
		parents,
		author,
		committer,
		option.Message)

	repo := core.GetRepository()
	g := c.ToGitObject()
	c.SetId(g.Hash())

	repo.Put(c.Id(), g)

	return c.Id(), nil
}
