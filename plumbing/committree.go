package plumbing

import (
	"fmt"
	"io"
	"time"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

type CommitTreeOption struct {
	Parents []string
	Message string
}

// Reads tree information into the index.
func CommitTree(w io.Writer, tree common.Hash, option *CommitTreeOption) error {
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
	parents := make([]common.Hash, len(option.Parents))
	for i, p := range option.Parents {
		parents[i], _ = common.NewHash(p)
	}

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

	fmt.Fprintf(w, "%s\n", c.Id())
	return nil
}
