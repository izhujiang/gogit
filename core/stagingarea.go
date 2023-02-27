package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"

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

func (s *StagingArea) Stage(path string) error {
	panic("Not implemented")
}

func (s *StagingArea) Unstage(path string) {
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
	gObj, err := repo.Get(treeId)
	if err != nil {
		return err
	}
	if gObj.Type() != object.ObjectTypeTree {
		return ErrIsNotATreeObject
	}

	tree := object.GitObjectToTree(gObj)
	tree.ShowContent(os.Stdout)
	// TODO: Add tree object to Index

	// return idx.SaveIndex(s.path)
	return nil
}

// func (s *StagingArea) readTree(repo *Repository, treeId common.Hash, prefix string) []*index.IndexEntry {
// 	gObj, _ := repo.Get(treeId)
// 	tree := GitObjectToTree(gObj)

// 	idx_entries := make([]*index.IndexEntry, 0)
// 	for _, entry := range tree.entries {
// 		switch entry.Type {
// 		case ObjectTypeBlob:
// 			ie := index.NewIndexEntry(
// 				entry.Oid,
// 				entry.Mode,
// 				filepath.Join(prefix, entry.Name))
// 			idx_entries = append(idx_entries, ie)
// 		case ObjectTypeTree:
// 			s.readTree(repo, entry.Oid, filepath.Join(prefix, entry.Name))
// 		default:
// 			log.Fatal("Unknown Entry Type")
// 		}
// 	}
// 	return idx_entries

// }

// read .git/index file and using files to build and save trees
func (s *StagingArea) WriteTree() (common.Hash, error) {
	idx := index.LoadIndex(s.path)

	repo := GetRepository()
	var saveTree = func(t *object.Tree) {
		g := object.TreeToGitObject(t)
		t.SetId(g.Hash())

		repo.Put(t.Id(), g)
	}

	treeId, err := idx.WriteTree(saveTree)
	if err != nil {
		log.Fatal(err)
	}

	idx.SaveIndex(s.path)

	return treeId, nil
}

// UpdateIndexEntry add or replace IndexEntry identified by path, and Invalidate all entries in TreeCache covered by path
func (s *StagingArea) UpdateIndex(oid common.Hash, path string) {
	idx := index.LoadIndex(s.path)

	entry, _ := idx.FindIndexEntry(path)

	if entry == nil {
		entry = index.NewIndexEntry(oid, common.Regular, path)
		idx.InsertIndexEntry(entry)
	}

	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	stat := fi.Sys().(*syscall.Stat_t)
	entry.Oid = oid
	entry.Mode = common.FileMode(stat.Mode)
	entry.CTime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	entry.MTime = fi.ModTime()
	entry.Dev = uint32(stat.Dev)
	entry.Ino = uint32(stat.Ino)
	entry.Uid = stat.Gid
	entry.Gid = stat.Gid
	entry.Size = uint32(stat.Size)

	idx.SaveIndex(s.path)
}

func (s *StagingArea) UpdateIndexFromCache(oid common.Hash, path string, mode common.FileMode) {
	idx := index.LoadIndex(s.path)

	entry, _ := idx.FindIndexEntry(path)

	fmt.Println("found entry: ", entry)
	if entry == nil {
		entry = index.NewIndexEntry(oid, mode, path)
		idx.InsertIndexEntry(entry)
	} else {
		entry.Oid = oid
		entry.Mode = mode
		entry.CTime = time.Unix(0, 0)
		entry.MTime = time.Unix(0, 0)
		entry.Dev = 0
		entry.Ino = 0
		entry.Uid = 0
		entry.Gid = 0
		entry.Size = 0
		entry.IntentToAdd = false
		entry.Skipworktree = false
		entry.Stage = 0
		idx.InvalidatePathInCacheTree(path)
	}

	idx.SaveIndex(s.path)
}

// If a specified file is in the index but is missing then it’s removed. Default behavior is to ignore removed file.
func (s *StagingArea) UpdateIndexRemove(path string) {
	idx := index.LoadIndex(s.path)

	idx.RemoveIndexEntry(path)
	idx.SaveIndex(s.path)
}
