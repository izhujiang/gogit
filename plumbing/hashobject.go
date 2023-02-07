package plumbing

import (
	"bytes"
	"io"
	"log"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
)

func HashObject(r io.Reader, t core.ObjectType) (common.Hash, *core.GitObject, error) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		log.Fatal(err)
	}
	g := core.NewGitObject(t, buf.Bytes())
	h := g.Hash()

	return h, g, err
}
