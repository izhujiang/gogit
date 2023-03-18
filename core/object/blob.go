package object

import (
	"io"

	"github.com/izhujiang/gogit/common"
)

// Blob object, implements Interface TreeEntry{ Object, fs.DirEntry}
type Blob struct {
	GitObject
	name     string
	filemode common.FileMode
}

func EmptyBlob() *Blob {
	return &Blob{
		GitObject: GitObject{
			objectKind: Kind_Blob,
		},
	}
}

// func EmptyBlobWithId(oid common.Hash) *Blob {
// 	return &Blob{
// 		GitObject: GitObject{
// 			oid:        oid,
// 			objectKind: Kind_Blob,
// 		},
// 	}
// }

func NewBlob(h common.Hash, name string, filemode common.FileMode, content []byte) *Blob {
	b := &Blob{
		GitObject: GitObject{
			oid:        h,
			objectKind: Kind_Blob,
			content:    content,
		},
		name:     name,
		filemode: filemode,
	}

	return b
}

func (b *Blob) Name() string {
	// return filepath.Base(b.fullpath)
	return b.name
}

func (b *Blob) IsDir() bool {
	return false
}
func (b *Blob) Type() common.FileMode {
	return common.Regular
}

func (b *Blob) Info() (common.FileInfo, error) {
	return &GitObjectInfo{}, nil
}

// type File interface {
// 	Stat() (FileInfo, error)
// 	Read([]byte) (int, error)
// 	Close() error
// }

// implement fs.File
// func (b *Blob) Stat() (fs.FileInfo, error) {
// 	return &GitObjectInfo{}, nil
// }

// func (b *Blob) Read(buf []byte) (int, error) {
// 	if b.pos >= len(buf) {
// 		return 0, io.EOF
// 	} else if b.pos+len(buf) < len(b.content) {
// 		n := copy(buf, b.content)
// 		return n, nil
// 	} else {
// 		n := copy(buf, b.content)
// 		return n, io.EOF
// 	}

// }

// func (b *Blob) Close() error {
// 	b.pos = len(b.content)
// 	return nil
// }

//	func (b *Blob) SetId(oid common.Hash) {
//		b.oid = oid
//	}

func (b *Blob) SetName(name string) {
	// return filepath.Base(b.fullpath)
	b.name = name
}

func (b *Blob) FromGitObject(g *GitObject) {
	b.GitObject = *g
	b.parseContent()
}

func GitObjectToBlob(g *GitObject) *Blob {
	b := &Blob{
		GitObject: *g,
	}

	b.parseContent()

	return b
}

func (b *Blob) Serialize(w io.Writer) error {
	b.composeContent()

	return b.GitObject.Serialize(w)
}

func (b *Blob) Deserialize(r io.Reader) error {
	err := b.GitObject.Deserialize(r)

	if err != nil {
		return err
	}

	b.parseContent()
	return nil
}

func (b *Blob) parseContent() {
	// Do nothing
}

func (b *Blob) composeContent() {
	// Do nothing
}
