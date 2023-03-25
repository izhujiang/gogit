package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/internal/utils"
	"github.com/izhujiang/gogit/core/object"
)

var (
	errObjectNotExists     = errors.New("Object does not exist.")
	errRepositoryNotExists = errors.New("Git repository does not exist, which should be initialized.")
)

type Repository struct {
	Name string
	// relative to the root of workspace
	Path string
}

// Init Git Repository in the path. Default, root == "."
func (r *Repository) InitRepository(w io.Writer, root string) error {
	if root == "" {
		root = "./"
	} else {
		os.MkdirAll(root, 0755)
	}

	setupRepositoryFramework(w, filepath.Join(root, r.Path))

	return nil
}

func (r *Repository) Put(g *object.GitObject) error {
	soid := g.Id().String()

	dir := filepath.Join(r.ObjectsPath(), soid[:2])
	path := filepath.Join(dir, soid[2:])

	if g == nil || utils.FileExists(path) {
		return nil
	}

	os.MkdirAll(dir, 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = g.Save(f)
	return err
}

// Get GitObject from Repository, return nil and error if oid is invalide
func (r *Repository) Get(oid common.Hash) (*object.GitObject, error) {
	path, err := r.checkObjectExists(oid)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	g, err := object.Load(f, oid)
	return g, err
}

func (r *Repository) GetAsBlob(oid common.Hash) (*object.Blob, error) {
	if oid == common.ZeroHash {
		return nil, errObjectNotExists
	}

	g, err := r.Get(oid)
	if err != nil {
		return nil, err
	}
	b := object.GitObjectToBlob(g)
	return b, nil
}

func (r *Repository) GetAsTree(oid common.Hash) (*object.Tree, error) {
	if oid == common.ZeroHash {
		return nil, errObjectNotExists
	}

	g, err := r.Get(oid)
	if err != nil {
		return nil, err
	}
	t := object.GitObjectToTree(g)
	return t, nil
}

// Load multiple Trees led by rootId from repository
func (r *Repository) LoadTrees(rootId common.Hash) (*object.Tree, error) {
	root, err := r.GetAsTree(rootId)
	if err != nil {
		return nil, err
	}

	tq := object.NewQueue()
	tq.Enqueue(root)
	for {
		t := tq.Dequeue()
		if t == nil {
			break
		}

		t.ForEach(func(e *object.TreeEntry) error {
			if e.Kind == object.Kind_Tree {
				sub_t, err := r.GetAsTree(e.Oid)
				if err == nil {
					e.Pointer = sub_t
					tq.Enqueue(sub_t)
				}
			}

			return nil
		})
	}

	return root, nil
}

func (r *Repository) Dump(oid common.Hash, w io.Writer) error {
	path, err := r.checkObjectExists(oid)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	object.DumpGitObject(f, w)
	return err
}

func (r *Repository) checkObjectExists(oid common.Hash) (string, error) {
	name := oid.String()
	dir := filepath.Join(r.ObjectsPath(), name[:2])
	path := filepath.Join(dir, name[2:])

	if !utils.FileExists(path) {
		return "", errObjectNotExists
	}

	return path, nil
}

func (r *Repository) ObjectsPath() string {
	return filepath.Join(r.Path, "objects")
}

// --------------------------------------------------------------------------
func setupRepositoryFramework(w io.Writer, path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		os.MkdirAll(path, 0755)

		os.Mkdir(filepath.Join(path, "hooks"), 0755)
		os.Mkdir(filepath.Join(path, "info"), 0755)
		os.Mkdir(filepath.Join(path, "objects"), 0755)
		os.Mkdir(filepath.Join(path, "objects", "info"), 0755)
		os.Mkdir(filepath.Join(path, "objects", "pack"), 0755)
		os.Mkdir(filepath.Join(path, "refs"), 0755)
		os.Mkdir(filepath.Join(path, "refs", "heads"), 0755)
		os.Mkdir(filepath.Join(path, "refs", "tags"), 0755)

		head := "ref: refs/heads/main"
		os.WriteFile(filepath.Join(path, "HEAD"), []byte(head), 0644)

		config := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
	ignorecase = true
	precomposeunicode = true`
		os.WriteFile(filepath.Join(path, "config"), []byte(config), 0644)

		desc := "Unnamed repository; edit this file 'description' to name the repository."
		os.WriteFile(filepath.Join(path, "description"), []byte(desc), 0644)

		path, _ = filepath.Abs(path)
		msg := fmt.Sprintf("Initialized empty Git repository in %s/\n", path)
		w.Write([]byte(msg))

	} else {
		path, _ = filepath.Abs(path)
		log.Printf("Git repository has existed in  %s/", path)
	}
}
