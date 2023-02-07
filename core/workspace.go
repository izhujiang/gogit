package core

import (
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
	WorkingArea  *WorkingArea
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
				WorkingArea: &WorkingArea{},
			}
		}
	}
	return singleInstance, nil
}

// TODO: log and return err
func (w *Workspace) InitWorkspace(root string) {
	if root == "" {
		root = "./"
	} else {
		os.MkdirAll(root, 0755)
	}
	w.repository.InitRepository(root)

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
