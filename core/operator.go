package core

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/object"
)

func HashObjectFromPath(path string, oType object.ObjectType, save bool) (common.Hash, error) {
	f, err := os.Open(path)
	if err != nil {
		return common.ZeroHash, err
	}
	defer f.Close()

	return HashObjectFromReader(f, oType, save)
}

func HashObjectFromReader(r io.Reader, oType object.ObjectType, save bool) (common.Hash, error) {
	h, gObj, err := hashObject(r, oType)

	if save {
		repo := GetRepository()
		// fmt.Printf("obj = %+v\n", gObj)
		repo.Put(h, gObj)
	}

	return h, err
}

func hashObject(r io.Reader, t object.ObjectType) (common.Hash, *object.GitObject, error) {
	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(r)
	if err != nil {
		log.Fatal(err)
	}
	g := object.NewGitObject(t, buf.Bytes())
	h := g.Hash()

	return h, g, err
}
