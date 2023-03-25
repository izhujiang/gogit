package core

import (
	"io"
	"os"
	"path/filepath"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/internal/index"
	"github.com/izhujiang/gogit/core/object"
)

type StagingArea struct {
	path string
	index.Index
}

func (s *StagingArea) Load() {
	// s.Index.Load(s.path)
	idx := &s.Index
	idx.Load(s.path)
}

func (s *StagingArea) Save() error {
	idx := &s.Index
	return idx.Save(s.path)
	// return s.Index.Save(s.path)
}

func (s *StagingArea) Stage(paths []string) error {
	idx := &s.Index

	for _, fp := range paths {
		s.updateIndex(idx, fp)
	}

	idx.Sort()

	return nil
}
func (s *StagingArea) Unstage(paths []string, recursive bool) {
	idx := &s.Index
	for _, fp := range paths {
		idx.Remove(fp, recursive)
	}

}

// func (s *StagingArea) LsFiles(w io.Writer, withDetail bool) {

// 	s.ListIndex(w, withDetail)
// }

// Reads tree information into the index
func (s *StagingArea) ReadTree(treeId common.Hash, prefix string, eraseOriginal bool) error {
	idx := &s.Index
	if eraseOriginal == true {
		idx.Reset()
	}
	repo := GetRepository()

	// load trees led by root from repo and add to CacheTree
	root, err := repo.LoadTrees(treeId)
	if err != nil {
		return err
	}
	root = updateRootWithPrefix(root, prefix)
	fs := object.NewTreeFs(root)
	idx.ReadTrees(fs)

	// Update Cachetree and save to repository
	idx.UpdateCacheTree()
	idx.CacheTree.DFWalk(func(path string, t *object.Tree) error {
		t.RegularizeEntries()

		if t.Id() == common.ZeroHash {
			t.Sort()
			t.Hash()

			g := t.ToGitObject()
			repo.Put(g)
		}

		return nil
	}, false)

	return err
}

func updateRootWithPrefix(root *object.Tree, prefix string) *object.Tree {
	if prefix == "." || prefix == "" {
		return root
	}

	path := prefix
	// build trees from prefix and link trees from repo
	for {
		base := filepath.Base(path)
		if base == "." {
			break
		}

		// parent tree
		pTree := object.EmptyTree()

		entry := object.NewTreeEntry(root.Id(), base, common.Dir)
		entry.Pointer = root
		pTree.Append(entry)

		root = pTree

		path = filepath.Dir(path)
	}

	return root
}

// read .git/index file and using files to build and save trees
func (s *StagingArea) WriteTree() (common.Hash, error) {
	idx := &s.Index
	repo := GetRepository()

	idx.UpdateCacheTree()
	idx.CacheTree.DFWalk(func(path string, t *object.Tree) error {
		t.RegularizeEntries()

		if t.Id() == common.ZeroHash {
			t.Sort()
			t.Hash()
			g := t.ToGitObject()
			repo.Put(g)
		}

		return nil
	}, false)

	return idx.CacheTree.Root().Id(), nil
}

func (s *StagingArea) updateIndex(idx *index.Index, path string) error {
	e := idx.Find(path)
	fi, _ := os.Stat(path)

	// file has not existed in idx of has been modified
	if e == nil {
		oid, err := HashObjectFromPath(path, object.Kind_Blob, true)
		if err != nil {
			return err
		}

		e = index.NewIndexEntryWithFileInfo(oid, common.Regular, path, fi)
		idx.Append(e)

	} else if e.ModTime().Before(fi.ModTime()) {
		oid, err := HashObjectFromPath(path, object.Kind_Blob, true)
		if err != nil {
			return err
		}

		idx.Update(e, oid, fi)
		// e.Update(oid, fi)
	}

	return nil
}

// UpdateIndexEntry add or replace IndexEntry identified by path, and Invalidate all entries in TreeCache covered by path
func (s *StagingArea) UpdateIndex(path string) {
	idx := &s.Index
	s.updateIndex(idx, path)
	idx.Sort()
}

func (s *StagingArea) UpdateIndexFromCache(oid common.Hash, path string, mode common.FileMode) {
	idx := &s.Index

	// file has not existed in idx of has been modified
	e := idx.Find(path)
	if e == nil {
		e = index.NewIndexEntry(oid, common.Regular, path)
		idx.Append(e)
		idx.Sort()
	} else {
		idx.Update(e, oid, nil)
		idx.Sort()
	}
}

// If a specified file is in the index but is missing then it’s removed. Default behavior is to ignore removed file.
func (s *StagingArea) UpdateIndexRemove(path string) {
	idx := &s.Index

	idx.Remove(path, false)
}

func (s *StagingArea) Dump(w io.Writer) {
	idx := &s.Index
	idx.Dump(w)
}
