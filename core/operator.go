package core

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/object"
)

func HashObjectFromPath(path string, oType object.ObjectKind, save bool) (common.Hash, error) {
	f, err := os.Open(path)
	if err != nil {
		return common.ZeroHash, err
	}
	defer f.Close()

	return HashObjectFromReader(f, oType, save)
}

func HashObjectFromReader(r io.Reader, oType object.ObjectKind, save bool) (common.Hash, error) {
	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(r)
	if err != nil {
		log.Fatal(err)
	}

	var h common.Hash
	if save == false {
		h = common.HashObject(oType.String(), buf.Bytes())
	} else {
		repo := GetRepository()
		g := object.NewGitObject(oType, buf.Bytes())
		repo.Put(g)
		h = g.Id()
	}

	return h, err
}
