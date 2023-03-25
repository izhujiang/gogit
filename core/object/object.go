// don't do go generate ./...  //go:generate stringer -type=ObjectType
package object

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/izhujiang/gogit/common"
)

type ObjectKind byte

const (
	Kind_Blob ObjectKind = iota
	Kind_Tree
	Kind_Commit
	Kind_Tag
	Kind_Unknow
)

var (
	ErrGitObjectDataCorruptted = errors.New("Git object data corruptted.")
	ErrInvalidObject           = errors.New("Invalidate Object.")
)

func (t ObjectKind) String() string {
	switch t {
	case Kind_Blob:
		return "blob"
	case Kind_Tree:
		return "tree"
	case Kind_Commit:
		return "commit"
	case Kind_Tag:
		return "tag"
	default:
		// TODO: Others object type should be implemented.
		return "unknown"
	}
}

func ParseObjectKind(kind string) ObjectKind {
	switch kind {
	case "blob":
		return Kind_Blob
	case "tree":
		return Kind_Tree
	case "commit":
		return Kind_Commit
	case "tag":
		return Kind_Tag
	default:
		return Kind_Unknow
	}
}
func ObjectKindFromFilemode(fm common.FileMode) ObjectKind {
	if common.IsFile(fm) {
		return Kind_Blob
	} else if common.IsDir(fm) {
		return Kind_Tree
	} else {
		return Kind_Unknow
	}
}

type Object interface {
	Id() common.Hash
	Kind() ObjectKind

	Content() string
}

// GitObject, unmodifiable object
type GitObject struct {
	oid common.Hash
	// header
	objectKind ObjectKind

	content []byte // unzipped content
}

func NewGitObject(t ObjectKind, content []byte) *GitObject {
	s := len(content)
	c := make([]byte, s)
	copy(c, content)

	g := &GitObject{
		objectKind: t,
		content:    c,
	}

	// TODO: add condition to control hash or not.
	g.Hash()

	return g
}

func (g *GitObject) Id() common.Hash {
	return g.oid
}

func (g *GitObject) Kind() ObjectKind {
	return g.objectKind
}

func (g *GitObject) Size() int64 {
	return int64(len(g.content))
}

func (g *GitObject) Content() string {
	return string(g.content)
}

func (g *GitObject) Hash() common.Hash {
	g.oid = common.HashObject(g.objectKind.String(), g.content)
	return g.oid
}

// Load GitObject from stream (git repository)
func Load(r io.Reader, oid common.Hash) (*GitObject, error) {
	zr, _ := zlib.NewReader(r)
	defer zr.Close()

	var objtype string
	var size int64
	// read header
	fmt.Fscanf(zr, "%s %d\x00", &objtype, &size)

	kind := ParseObjectKind(objtype)
	content := make([]byte, size)
	n, _ := zr.Read(content)
	g := NewGitObject(kind, content)

	if n != len(g.content) || oid != g.Id() {
		return nil, ErrGitObjectDataCorruptted
	}

	return g, nil
}

// Save GitObject to stream (git repository)
func (g *GitObject) Save(w io.Writer) error {
	wt := zlib.NewWriter(w)
	defer wt.Close()

	// write header and content
	size := int64(len(g.content))
	header := fmt.Sprintf("%s %d\x00", strings.ToLower(g.objectKind.String()), size)
	_, err := wt.Write([]byte(header))
	_, err = wt.Write(g.content)

	err = wt.Flush()

	return err
}

// Dump object in .git repository
func DumpGitObject(r io.Reader, w io.Writer) {
	buf := &bytes.Buffer{}
	reader, _ := zlib.NewReader(r)
	defer reader.Close()

	io.Copy(buf, reader)

	// fmt.Println(string(buf.Bytes()))

	b := make([]byte, 16)
	empty := make([]byte, 16)
	addr := 0x00
	for {
		n, err := buf.Read(b)
		if err == io.EOF {
			break
		}

		if n < 16 {
			copy(b[n:], empty[n:])

		}

		bb := strings.ReplaceAll(string(b), "\t", "\\t")
		bb = strings.ReplaceAll(bb, "\n", "\\n")
		fmt.Fprintf(w, "%08x  % x  % x  %s\n", addr, b[:8], b[8:], bb)
		addr += 16
	}
}
