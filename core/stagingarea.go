package core

import (
	"bytes"
	"io"
	"log"
	"os"
)

type StagingArea struct {
}

var saInstance = &StagingArea{}

func GetStagingArea() *StagingArea {
	return saInstance
}

func (s *StagingArea) Stage(path string) error {
	repo, err := GetRepository()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	buf := &bytes.Buffer{}
	io.Copy(buf, f)
	obj, _ := HashObject(buf.Bytes(), ObjectTypeBlob)
	err = repo.Put(obj)
	return err
}

func (s *StagingArea) Unstage(path string) {

}
