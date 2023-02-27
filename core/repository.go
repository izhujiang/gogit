package core

import (
	"errors"
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
func (r *Repository) InitRepository(root string) error {
	log.Println("init gogit repository.")
	if root == "" {
		root = "./"
	} else {
		os.MkdirAll(root, 0755)
	}

	setupRepositoryFramework(filepath.Join(root, r.Path))

	return nil
}

func (r *Repository) Put(h common.Hash, obj *object.GitObject) error {
	oid := h.String()
	dir := filepath.Join(r.ObjectsPath(), oid[:2])
	path := filepath.Join(dir, oid[2:])

	if utils.FileExists(path) {
		return nil
		// log.Fatal("git object has existed: ", oid)
	}

	os.MkdirAll(dir, 0755)
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer f.Close()

	err = obj.Serialize(f)
	return err
}

func (r *Repository) Get(oid common.Hash) (*object.GitObject, error) {
	path, err := r.checkObjectExists(oid)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println("object path:", path)
	// TODO: handle reading from git repository (blob, tree, commmit and tag)
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer f.Close()

	obj, err := object.DeserializeGitObject(f)
	return obj, err
}

func (r *Repository) Dump(oid common.Hash, w io.Writer) error {
	path, err := r.checkObjectExists(oid)
	if err != nil {
		return err
	}

	// fmt.Println(path)
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

// func GetRepository() (*Repository, error) {
// 	if singleInstance == nil {
// 		lock.Lock()
// 		defer lock.Unlock()

// 		if !utils.DirectoryExists(repositoryRoot) {
// 			return nil, errRepositoryNotExists
// 		}

// 		if singleInstance == nil {
// 			singleInstance = &Repository{Name: repositoryName, Path: repositoryRoot}
// 		}
// 	}

// 	return singleInstance, nil
// }

func (r *Repository) ObjectsPath() string {
	return filepath.Join(r.Path, "objects")
}

// --------------------------------------------------------------------------
// internal functions

func setupRepositoryFramework(path string) error {
	foldInfo, err := os.Stat(path)
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

	} else {
		log.Println(foldInfo)
	}

	// TODO: return non-nil error when fail to create .gogit repository in phisical device.
	return nil
}
