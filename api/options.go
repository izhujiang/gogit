package git

import "github.com/izhujiang/gogit/plumbing"

// export git types.

type ObjectType string

const (
	ObjectTypeBlob   ObjectType = "blob"
	ObjectTypeTree   ObjectType = "tree"
	ObjectTypeCommit ObjectType = "commit"
	ObjectTypeTag    ObjectType = "tree"
)

type HashObjectOption struct {
	ObjectType ObjectType
	Write      bool
}

type CatFileOption plumbing.CatFileOption
type DumpObjectOption plumbing.DumpObjectOption
type LsTreeOption plumbing.LsTreeOption

type AddOption struct {
}

type LsFilesOption plumbing.LsFilesOption
