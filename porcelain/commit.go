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
	// TODO: do nothing if nothing new in the StagingArea
	// if (nothing new){
	// return nil
	// }

	wto := &plumbing.WriteTreeOption{}
	treeId, err := plumbing.WriteTree(wto)

	if err != nil { // fail to WriteTree, including trees in cache are already valid
		// TODO: promote as git status
		fmt.Println(err, treeId)
		return err

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
		repo := core.GetRepository()
		g, _ := repo.Get(lastCommitId)
		var changes *common.Changes
		if g != nil {

			lastCommit := object.GitObjectToCommit(g)
			lastTreeId := lastCommit.Tree()

			changes, err = compareTrees(lastTreeId, treeId)
		} else {
			changes, err = compareTrees(common.ZeroHash, treeId)
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

	thisTree, err := repo.LoadTrees(thisTreeId)
	if err != nil {
		return nil, err
	}
	thisTrees := object.NewTreeFs(thisTree)
	thisTreeCollector := &filesCollector{}
	thisTrees.DFWalk(thisTreeCollector.collect, true)

	lastTree, err := repo.LoadTrees(lastTreeId)
	var changes *common.Changes
	if err != nil || lastTree == nil {
		lastTreePairs := common.NameHashPairs(make([]*common.NameHashPair, 0))

		changes = common.CompareOrderedNameHashPairs(lastTreePairs, thisTreeCollector.pairs)
	} else {
		lastTrees := object.NewTreeFs(lastTree)
		lastTreeCollector := &filesCollector{}
		lastTrees.DFWalk(lastTreeCollector.collect, true)

		changes = common.CompareOrderedNameHashPairs(lastTreeCollector.pairs, thisTreeCollector.pairs)
	}

	return changes, nil

}

type filesCollector struct {
	pairs common.NameHashPairs
}

func (fc *filesCollector) collect(path string, t *object.Tree) error {
	t.ForEach(func(e *object.TreeEntry) error {
		if e.Kind == object.Kind_Blob {
			fp := filepath.Join(path, e.Name)
			p := &common.NameHashPair{
				Name: fp,
				Oid:  e.Oid,
				Mode: e.Filemode,
			}
			fc.pairs = append(fc.pairs, p)
		}

		return nil
	})

	return nil
}
