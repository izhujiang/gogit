package porcelain

import (
	"fmt"
	"io"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
	"github.com/izhujiang/gogit/plumbing"
)

type CommitOption struct {
	Message string
}

func Commit(w io.Writer, option *CommitOption) error {
	wto := &plumbing.WriteTreeOption{}
	treeId, err := plumbing.WriteTree(wto)

	if err != nil { // fail to WriteTree, including trees in cache are already valid
		// TODO: promote as git status

	} else {
		refs := core.GetReferencs()
		lastCommitId, err := refs.LastCommit()

		parents := make([]common.Hash, 0)
		if err == nil {
			parents = append(parents, lastCommitId)
		}

		// commit changes into repository
		ctOption := &plumbing.CommitTreeOption{
			Parents: parents,
			Message: option.Message,
		}
		commitId, err := plumbing.CommitTree(treeId, ctOption)
		if err == nil {
			// save commit id to ref/head/{branch}
			ref := core.GetReferencs()
			err = ref.SaveCommit(commitId)
			headMsg := fmt.Sprintf("[%s %s] %s\n", refs.Head(), commitId, ctOption.Message)
			w.Write([]byte(headMsg))
		}

		// TODO: list all commited files

		lastCommit := object.EmptyCommit()
		repo := core.GetRepository()
		gObj, _ := repo.Get(lastCommitId)
		lastCommit.FromGitObject(gObj)
		lastTreeId := lastCommit.Tree()

		changes := compareTrees(lastTreeId, treeId)

		for _, c := range changes.Create {
			line := fmt.Sprintf("create mode %s %s\n", c.To.Mode, c.To.Name)
			w.Write([]byte(line))
			// TODO: stat lines inserted
		}
		for _, c := range changes.Remove {
			line := fmt.Sprintf("delete mode %s %s\n", c.From.Mode, c.From.Name)
			w.Write([]byte(line))
			// TODO: stat lines deleted
		}

		// TODO: stat lines changed
		// for _, c := range changes.Modify {
		// 	line := fmt.Sprintf("delete mode %s %s\n", c.From.Mode, c.From.Name)
		// 	w.Write([]byte(line))
		// }
	}

	return err
}

func compareTrees(lastTreeId common.Hash, thisTreeId common.Hash) *common.Changes {
	repo := core.GetRepository()

	fillTree := func(t *object.Tree) {
		gObj, err := repo.Get(t.Id())
		if err != nil {
			return
		} else { // read content for tree identified by id
			if gObj.Type() != object.ObjectTypeTree {
				return
			}
			t.FromGitObject(gObj)
		}
	}

	lastTree := object.NewTree(lastTreeId, "")
	lastTrees := object.NewTreeCollection()
	lastTrees.InitWithRoot(lastTree)
	lastTrees.Expand(fillTree)

	lastTreeCollector := &filesCollector{}
	lastTrees.WalkByAlphabeticalOrder(lastTreeCollector.collect)

	thisTree := object.NewTree(thisTreeId, "")
	thisTrees := object.NewTreeCollection()
	thisTrees.InitWithRoot(thisTree)
	thisTrees.Expand(fillTree)

	thisTreeCollector := &filesCollector{}
	thisTrees.WalkByAlphabeticalOrder(thisTreeCollector.collect)

	changes := common.CompareOrderedNameHashPairs(lastTreeCollector.pairs, thisTreeCollector.pairs)
	return changes

}

type filesCollector struct {
	pairs common.NameHashPairs
}

func (fc *filesCollector) collect(path string, e object.TreeEntry) {
	if e.Type() == object.ObjectTypeBlob {
		p := &common.NameHashPair{
			Name: path,
			Oid:  e.Id(),
			Mode: e.Mode(),
		}
		fc.pairs = append(fc.pairs, p)
	}

}
