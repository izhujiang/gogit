package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

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
		if e == nil || e.ModTime().Before(fi.ModTime()) {
			oid, err := HashObjectFromPath(path, object.Kind_Blob, true)
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
	idx.LsIndexEntries(w, withDetail)
}

// Reads tree information into the index
func (s *StagingArea) ReadTree(treeId common.Hash, prefix string, eraseOriginal bool) error {
	idx := index.LoadIndex(s.path)

	if eraseOriginal == true {
		idx.Reset()
	}

	repo := GetRepository()
	// load trees from repo
	base := filepath.Base(prefix)
	root, err := repo.LoadTrees(treeId, base)
	if err != nil {
		return err
	}
	root = updateRootWithPrefix(root, prefix)
	fs := object.NewTreeFs(root)

	var saveTree = func(t *object.Tree) error {
		if t.Id() == common.ZeroHash {
			t.Hash()

			repo.Put(t.Id(), &t.GitObject)
		}
		return nil
	}

	err = idx.ReadTrees(fs, saveTree)
	if err != nil {
		return err
	}
	return idx.SaveIndex(s.path)
}

func updateRootWithPrefix(root *object.Tree, prefix string) *object.Tree {
	if prefix == "." || prefix == "" {
		return root
	}

	path := filepath.Dir(prefix)
	// build trees from prefix and link trees from repo
	for {
		base := filepath.Base(path)
		if base == "." {
			base = ""
		}

		pTree := object.NewTree(common.ZeroHash, base, common.Dir)
		pTree.UpdateOrAddEntry(root)
		pTree.Sort()
		// pTree.Hash()
		root = pTree

		if base == "" {
			break
		}

		path = filepath.Dir(path)
	}

	return root
}

// read .git/index file and using files to build and save trees
func (s *StagingArea) WriteTree() (common.Hash, error) {
	idx := index.LoadIndex(s.path)

	repo := GetRepository()

	treeId, err := idx.WriteTree(func(t *object.Tree) error {
		oid := t.Hash()

		repo.Put(oid, &t.GitObject)
		return nil
	})
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

	// idx.InvalidatePathInCacheTree(path)
	idx.SaveIndex(s.path)
}

func (s *StagingArea) UpdateIndexFromCache(oid common.Hash, path string, mode common.FileMode) {
	idx := index.LoadIndex(s.path)

	entry := index.NewIndexEntry(oid, mode, path)
	idx.UpdateOrInsertIndexEntry(entry)

	// idx.InvalidatePathInCacheTree(path)

	idx.SaveIndex(s.path)
}

// If a specified file is in the index but is missing then it’s removed. Default behavior is to ignore removed file.
func (s *StagingArea) UpdateIndexRemove(path string) {
	idx := index.LoadIndex(s.path)

	idx.RemoveIndexEntry(path)
	idx.SaveIndex(s.path)
}
