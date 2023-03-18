package porcelain

import (
	"fmt"
	"io"
	"path/filepath"

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

		changes, err := compareTrees(lastTreeId, treeId)
		if err != nil {
			return err
		}

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

func compareTrees(lastTreeId common.Hash, thisTreeId common.Hash) (*common.Changes, error) {
	repo := core.GetRepository()

	lastTree, err := repo.LoadTrees(lastTreeId, "")
	if err != nil {
		return nil, err
	}
	lastTrees := object.NewTreeFs(lastTree)

	lastTreeCollector := &filesCollector{}
	// lastTrees.WalkTreeEntryByAlphabeticalOrder(lastTreeCollector.collect)
	lastTrees.DFWalk(lastTreeCollector.collect, true)

	thisTree, err := repo.LoadTrees(thisTreeId, "")
	if err != nil {
		return nil, err
	}
	thisTrees := object.NewTreeFs(thisTree)

	thisTreeCollector := &filesCollector{}
	thisTrees.DFWalk(thisTreeCollector.collect, true)

	changes := common.CompareOrderedNameHashPairs(lastTreeCollector.pairs, thisTreeCollector.pairs)
	return changes, nil

}

type filesCollector struct {
	pairs common.NameHashPairs
}

func (fc *filesCollector) collect(path string, t *object.Tree) error {
	t.ForEach(func(e object.TreeEntry) error {
		if e.Kind() == object.Kind_Blob {
			fp := filepath.Join(path, e.Name())
			p := &common.NameHashPair{
				Name: fp,
				Oid:  e.Id(),
				Mode: e.Type(),
			}
			fc.pairs = append(fc.pairs, p)
		}

		return nil
	})

	return nil
}
