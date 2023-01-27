package porcelain

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/izhujiang/gogit/core"
)

func Add(path string) (core.Hash, error) {
	objType := core.ObjectTypeBlob

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buf := &bytes.Buffer{}
	io.Copy(buf, f)
	obj, err := core.HashObject(buf.Bytes(), objType)
	if err != nil {
		log.Fatal(err)
	}
	repo, err := core.GetRepository()
	if err != nil {
		log.Fatal(err)
	}

	err = repo.Put(obj)

	return obj.Id(), err

}
