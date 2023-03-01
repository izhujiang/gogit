package git

import (
	"github.com/izhujiang/gogit/plumbing"
	"github.com/izhujiang/gogit/porcelain"
)

// export git types.

type ObjectType string

const (
	ObjectTypeBlob   ObjectType = "blob"
	ObjectTypeTree   ObjectType = "tree"
	ObjectTypeCommit ObjectType = "commit"
	ObjectTypeTag    ObjectType = "tree"
)

type HashObjectOption struct {
	ObjectType string
	Write      bool
}
type CatFileOption plumbing.CatFileOption
type DumpOption plumbing.DumpOption
type LsTreeOption plumbing.LsTreeOption

type AddOption struct {
}

type LsFilesOption plumbing.LsFilesOption
type WriteTreeOption plumbing.WriteTreeOption
type ReadTreeOption plumbing.ReadTreeOption
type UpdateIndexOption plumbing.UpdateIndexOption
type CommitTreeOption plumbing.CommitTreeOption

type LogOption porcelain.LogOption
