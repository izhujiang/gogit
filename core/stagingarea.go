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
}

func (s *StagingArea) Stage(paths []string) error {
	idx := index.LoadIndex(s.path)

	for _, fp := range paths {
		s.updateIndex(idx, fp)
	}

	idx.Sort()

	idx.Save(s.path)
	return nil
}
func (s *StagingArea) Unstage(paths []string, recursive bool) {
	idx := index.LoadIndex(s.path)
	for _, fp := range paths {
		idx.Remove(fp, recursive)
	}

	idx.Save(s.path)
}

func (s *StagingArea) LsFiles(w io.Writer, withDetail bool) {
	idx := index.LoadIndex(s.path)
	idx.LsIndex(w, withDetail)
}

// Reads tree information into the index
func (s *StagingArea) ReadTree(treeId common.Hash, prefix string, eraseOriginal bool) error {
	idx := index.LoadIndex(s.path)

	if eraseOriginal == true {
		idx.Reset()
	}

	repo := GetRepository()
	// load trees from repo
	// base := filepath.Base(prefix)
	root, err := repo.LoadTrees(treeId)

	if err != nil {
		return err
	}
	root = updateRootWithPrefix(root, prefix)
	fs := object.NewTreeFs(root)

	var saveTree = func(t *object.Tree) error {
		if t.Id() == common.ZeroHash {
			t.Hash()
			g := t.ToGitObject()
			repo.Put(g)
		}
		return nil
	}

	err = idx.ReadTrees(fs, saveTree)

	if err != nil {
		return err
	}
	return idx.Save(s.path)
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
	idx := index.LoadIndex(s.path)

	repo := GetRepository()

	treeId, err := idx.WriteTree(func(t *object.Tree) error {
		if t.Id() == common.ZeroHash {
			t.Hash()

			g := t.ToGitObject()
			repo.Put(g)
		}
		return nil
	})

	if err != nil {
		return treeId, err
	}

	idx.Save(s.path)

	return treeId, nil
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
	idx := index.LoadIndex(s.path)

	s.updateIndex(idx, path)
	idx.Sort()

	// idx.InvalidatePathInCacheTree(path)
	idx.Save(s.path)
}

func (s *StagingArea) UpdateIndexFromCache(oid common.Hash, path string, mode common.FileMode) {
	idx := index.LoadIndex(s.path)

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

	idx.Save(s.path)
}

// If a specified file is in the index but is missing then it’s removed. Default behavior is to ignore removed file.
func (s *StagingArea) UpdateIndexRemove(path string) {
	idx := index.LoadIndex(s.path)

	idx.Remove(path, false)
	idx.Save(s.path)
}

func (s *StagingArea) Dump(w io.Writer) {
	idx := index.LoadIndex(s.path)
	idx.Dump(w)
}
