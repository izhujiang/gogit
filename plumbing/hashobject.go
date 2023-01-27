package plumbing

import (
	"bytes"
	"io"
	"log"

	"github.com/izhujiang/gogit/core"
)

func HashObject(r io.Reader, t core.ObjectType) (*core.GitObject, error) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		log.Fatal(err)
	}
	return core.HashObject(buf.Bytes(), t)
}
