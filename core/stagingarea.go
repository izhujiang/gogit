package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/internal/index"
	"github.com/izhujiang/gogit/core/object"
)

// inner state of StagingArea
type StagingArea struct {
	path string
}

var (
	ErrIsNotATreeObject = errors.New("Is not a valid tree object.")
)

func (s *StagingArea) Stage(filepaths []string) error {
	idx := index.LoadIndex(s.path)
	// repo := GetRepository()

	stage := func(path string) error {
		e := idx.FindIndexEntry(path)
		fi, _ := os.Stat(path)

		fmt.Println("staging ", path)
		// file has not existed in idx of has been modified
		if e == nil || e.MTime.Before(fi.ModTime()) {
			oid, err := HashObjectFromPath(path, object.ObjectTypeBlob, true)
			if err != nil {
				return err
			}

			nEntry := index.NewIndexEntryWithFileInfo(oid, common.Regular, path, fi)
			idx.UpdateOrInsertIndexEntry(nEntry)
		}

		return nil
	}

	for _, fp := range filepaths {
		stage(fp)
	}

	idx.SaveIndex(s.path)
	return nil
}

func (s *StagingArea) Unstage(paths []string) {
	panic("Not implemented")
}

func (s *StagingArea) Dump(w io.Writer) {
	idx := index.LoadIndex(s.path)
	idx.Dump(w)
}

func (s *StagingArea) LsFiles(w io.Writer, withDetail bool) {
	idx := index.LoadIndex(s.path)

	idx.ForeachIndexEntry(func(entry *index.IndexEntry) {
		if withDetail {
			fmt.Fprintf(w, "%o %s %d \t%s\n", entry.Mode, entry.Oid, entry.Stage, entry.Filepath)

		} else {
			fmt.Fprintln(w, entry.Filepath)
		}
	})

}

// Reads tree information into the index
func (s *StagingArea) ReadTree(treeId common.Hash, prefix string, eraseOriginal bool) error {
	idx := index.LoadIndex(s.path)

	if eraseOriginal == true {
		idx.Reset()
	}

	repo := GetRepository()

	trees := object.NewTreeCollection()
	trees.InitWithRootId(treeId, prefix)

	trees.Expand(func(t *object.Tree) {
		gObj, err := repo.Get(t.Id())
		if err != nil {
			return
		} else { // read content for tree identified by id
			if gObj.Type() != object.ObjectTypeTree {
				return
			}
			t.FromGitObject(gObj)
		}
	})

	var saveTree = func(t *object.Tree) {
		g := t.ToGitObject()
		t.SetId(g.Hash())

		repo.Put(t.Id(), g)
	}

	err := idx.ReadTrees(trees, saveTree)
	if err != nil {
		return err
	}
	return idx.SaveIndex(s.path)
}

// read .git/index file and using files to build and save trees
func (s *StagingArea) WriteTree() (common.Hash, error) {
	idx := index.LoadIndex(s.path)

	repo := GetRepository()
	var saveTree = func(t *object.Tree) {
		g := t.ToGitObject()
		t.SetId(g.Hash())

		repo.Put(t.Id(), g)
	}

	treeId, err := idx.WriteTree(saveTree)
	if err != nil {
		return treeId, err
	}

	idx.SaveIndex(s.path)

	return treeId, nil
}

// UpdateIndexEntry add or replace IndexEntry identified by path, and Invalidate all entries in TreeCache covered by path
func (s *StagingArea) UpdateIndex(oid common.Hash, path string) {
	idx := index.LoadIndex(s.path)

	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	entry := index.NewIndexEntryWithFileInfo(oid, common.Regular, path, fi)
	idx.UpdateOrInsertIndexEntry(entry)

	idx.Sort()

	idx.InvalidatePathInCacheTree(path)
	idx.SaveIndex(s.path)
}

func (s *StagingArea) UpdateIndexFromCache(oid common.Hash, path string, mode common.FileMode) {
	idx := index.LoadIndex(s.path)

	entry := index.NewIndexEntry(oid, mode, path)
	idx.UpdateOrInsertIndexEntry(entry)

	idx.InvalidatePathInCacheTree(path)

	idx.SaveIndex(s.path)
}

// If a specified file is in the index but is missing then it’s removed. Default behavior is to ignore removed file.
func (s *StagingArea) UpdateIndexRemove(path string) {
	idx := index.LoadIndex(s.path)

	idx.RemoveIndexEntry(path)
	idx.SaveIndex(s.path)
}
