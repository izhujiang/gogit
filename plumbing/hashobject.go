package plumbing

import (
	"bytes"
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

type HashObjectOption struct {
	ObjectType object.ObjectType
	Write      bool
}

func HashObject(r io.Reader, option *HashObjectOption) (common.Hash, error) {
	h, gObj, err := hashObject(r, option.ObjectType)

	if option.Write {
		repo := core.GetRepository()
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
