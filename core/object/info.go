package object

import (
	"io/fs"
	"time"
)

type GitObjectInfo struct {
}

func (g *GitObjectInfo) Name() string {
	panic("not implemented") // TODO: Implement
}

func (g *GitObjectInfo) Size() int64 {
	panic("not implemented") // TODO: Implement
}

func (g *GitObjectInfo) Mode() fs.FileMode {
	panic("not implemented") // TODO: Implement
}

func (g *GitObjectInfo) ModTime() time.Time {
	panic("not implemented") // TODO: Implement
}

func (g *GitObjectInfo) IsDir() bool {
	panic("not implemented") // TODO: Implement
}

func (g *GitObjectInfo) Sys() any {
	panic("not implemented") // TODO: Implement
}
