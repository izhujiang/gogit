package core

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/izhujiang/gogit/core/internal/utils"
)

// workspace include working area, staging area and git repository

const (
	repositoryName = ".git"
	repositoryRoot = ".git"
)

type Workspace struct {
	repository   *Repository
	stageingArea *StagingArea
	workingArea  *WorkingArea
	references   *References
}

var lock = &sync.Mutex{}
var singleInstance *Workspace

func GetWorkspace() (*Workspace, error) {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()

		if singleInstance == nil {
			singleInstance = &Workspace{
				repository: &Repository{Name: repositoryName, Path: repositoryRoot},
				stageingArea: &StagingArea{
					path: filepath.Join(repositoryRoot, "index"),
				},
				workingArea: &WorkingArea{},
				references: &References{
					root:     repositoryRoot,
					headpath: filepath.Join(repositoryRoot, "HEAD"),
				},
			}
		}
	}
	return singleInstance, nil
}

// TODO: log and return err
func (ws *Workspace) InitWorkspace(w io.Writer, root string) {
	if root == "" {
		root = "./"
	} else {
		os.MkdirAll(root, 0755)
	}
	ws.repository.InitRepository(w, root)

}

func GetRepository() *Repository {
	w, _ := GetWorkspace()
	if !utils.DirectoryExists(w.repository.Path) {
		log.Fatal(errRepositoryNotExists)
	}

	return w.repository
}

func GetStagingArea() *StagingArea {
	w, _ := GetWorkspace()
	return w.stageingArea
}

func GetReferencs() *References {
	w, _ := GetWorkspace()
	return w.references
}
