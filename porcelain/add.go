package porcelain

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

func Add(path string) (common.Hash, error) {
	objType := object.ObjectTypeBlob

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buf := &bytes.Buffer{}
	io.Copy(buf, f)
	gObj := object.NewGitObject(objType, buf.Bytes())
	h := gObj.Hash()

	// obj, err := core.HashObject(buf.Bytes(), objType)
	if err != nil {
		log.Fatal(err)
	}
	repo := core.GetRepository()

	err = repo.Put(h, gObj)

	return h, err

}
