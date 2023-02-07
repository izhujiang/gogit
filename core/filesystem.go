package core

import "github.com/izhujiang/gogit/common"

// Git stores content in a manner similar to a UNIX filesystem, but a bit simplified.
// All the content is stored as tree and blob objects, with trees corresponding to UNIX directory entries and blobs corresponding more or less to inodes or file contents.

type DirEntry interface {
	/* TODO: add methods */
	Id() common.Hash
	Type() ObjectType
	Name() string
	Size() int
	Content() []byte
}

type VirtualFileSystem struct {
	root DirEntry
}

func (fs *VirtualFileSystem) Mount(root string) {
	// fs := VirtualFileSystem{}
	panic("Not implemented")
}

func (fs *VirtualFileSystem) Unmount() {
	panic("Not implemented")
}

func (fs *VirtualFileSystem) Add(path string) {

}
