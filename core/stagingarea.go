package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/internal/filemode"
	"github.com/izhujiang/gogit/core/internal/index"
)

type treeDict map[string]*Tree

// inner state of StagingArea
type StagingArea struct {
	path string
}

func (s *StagingArea) Stage(path string) error {
	panic("Not implemented")
}

func (s *StagingArea) Unstage(path string) {
	panic("Not implemented")
}

func (s *StagingArea) Dump(w io.Writer) {
	f, err := os.Open(s.path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := &index.IndexDecoder{Reader: f}
	idx := index.New()
	decoder.Decode(idx)
	idx.Dump(w)
}

func (s *StagingArea) LsFiles(w io.Writer, withDetail bool) {
	f, err := os.Open(s.path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := &index.IndexDecoder{Reader: f}
	idx := &index.Index{}
	decoder.Decode(idx)

	for _, entry := range idx.Entries {
		if withDetail {
			fmt.Fprintf(w, "%o %s %d \t%s\n", entry.Mode, entry.Oid, entry.Stage, entry.Filepath)

		} else {
			fmt.Fprintln(w, entry.Filepath)
		}
	}

}

// Reads tree information into the index
func (s *StagingArea) ReadTree(treeId common.Hash, prefix string) {
	fmt.Println("read-tree:", treeId, prefix)
	// f, err := os.Open(s.path)
	f, err := os.OpenFile(s.path, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := &index.IndexDecoder{Reader: f}
	idx := &index.Index{}
	decoder.Decode(idx)

	if prefix == "" {
		idx.RemoveAll()
	}

	repo := GetRepository()
	// readTree := func() {

	// }
	idx_entries := make([]*index.IndexEntry, 0)
	idx_entries = append(idx_entries, s.readTree(repo, treeId, prefix)...)
	fmt.Printf("idx_entries = %+v\n", idx_entries)
	idx.InsertEntries(idx_entries)

	ff, err := os.OpenFile(s.path+".bak", os.O_CREATE|os.O_RDWR, 0644)
	// ff := os.OpenFile()
	// f.Seek(0, 0)
	encoder := &index.IndexEncoder{Writer: ff}
	encoder.Encode(idx)
}

func (s *StagingArea) readTree(repo *Repository, treeId common.Hash, prefix string) []*index.IndexEntry {
	gObj, _ := repo.Get(treeId)
	tree := NewTree(treeId, prefix)
	tree.FromGitObject(gObj)

	idx_entries := make([]*index.IndexEntry, 0)
	for _, entry := range tree.entries {
		switch entry.Type {
		case ObjectTypeBlob:
			ie := index.NewIndexEntry(
				entry.Oid,
				entry.Mode,
				filepath.Join(prefix, entry.Name))
			idx_entries = append(idx_entries, ie)
		case ObjectTypeTree:
			s.readTree(repo, entry.Oid, filepath.Join(prefix, entry.Name))
		default:
			log.Fatal("Unknown Entry Type")
		}
	}
	return idx_entries

}

// read .git/index file and using files to build and save trees
func (s *StagingArea) WriteTree() (common.Hash, error) {
	f, err := os.Open(s.path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := &index.IndexDecoder{Reader: f}
	idx := &index.Index{}
	decoder.Decode(idx)

	trees := treeCollection{}
	for _, entry := range idx.Entries {
		fmt.Printf("%o %s %d \t%s\n", entry.Mode, entry.Oid, entry.Stage, entry.Filepath)
		fp := entry.Filepath
		mode := filemode.FileMode(entry.Mode)
		ftId := entry.Oid
		trees.addFilePath(fp, mode, ObjectTypeBlob, ftId)
	}

	trees.DFWalk(normalize)

	trees.DFWalk(func(t *Tree) error {
		fmt.Printf("\t%s %s\n ", t.name, t.oid)
		for _, entry := range t.entries.sort() {
			fmt.Printf("\t%s %s\t%s\t%s\n ", entry.Mode, entry.Type, entry.Oid, entry.Name)
		}
		return nil
	})
	return trees["."].oid, err
}

// Hash the tree according to new entries and save to repository
func normalize(t *Tree) error {
	g := t.ToGitObject()
	t.oid = g.Hash()

	repo := GetRepository()
	err := repo.Put(t.oid, g)
	// fmt.Printf("writing %s\n", h)

	return err
}
