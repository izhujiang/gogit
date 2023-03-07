package plumbing

import (
	"io"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core"
	"github.com/izhujiang/gogit/core/object"
)

type HashObjectOption struct {
	ObjectType object.ObjectType
	Write      bool
}

func HashObject(r io.Reader, option *HashObjectOption) (common.Hash, error) {
	return core.HashObjectFromReader(r, option.ObjectType, option.Write)
}
