// don't do go generate ./...  //go:generate stringer -type=ObjectType
package core

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

type ObjectType byte

const (
	ObjectTypeBlob ObjectType = iota
	ObjectTypeTree
	ObjectTypeCommit
	ObjectTypeTag
	ObjectTypeUnknow
)

var (
	errInvalidObject = errors.New("Invalidate Object.")
)

func (t ObjectType) String() string {
	switch t {
	case ObjectTypeBlob:
		return "blob"
	case ObjectTypeTree:
		return "tree"
	case ObjectTypeCommit:
		return "commit"
	case ObjectTypeTag:
		return "tag"
	default:
		// TODO: Others object type should be implemented.
		return "unknown"
	}
}

var (
	objectTypeDict = map[string]ObjectType{"blob": ObjectTypeBlob,
		"tree":   ObjectTypeTree,
		"commit": ObjectTypeCommit,
		"tag":    ObjectTypeTag,
	}
)

func ParseObjectType(objType string) ObjectType {
	if ot, ok := objectTypeDict[strings.ToLower(objType)]; ok {
		return ot
	} else {
		return ObjectTypeUnknow
	}
}

// type Object interface {
// 	Id() Hash
// 	Type() ObjectType
// 	Size() int64
// 	Content() []byte
// }

type GitObject struct {
	id Hash
	// header
	objectType ObjectType
	size       int64

	content []byte // unzip content which
}

// Load GitObject from stream (git repository)
func LoadGitObject(oid Hash, r io.Reader) (*GitObject, error) {
	zr, _ := zlib.NewReader(r)
	defer zr.Close()

	var objtype string
	var size int64
	// read header
	fmt.Fscanf(zr, "%s %d\x00", &objtype, &size)

	g := &GitObject{
		objectType: ParseObjectType(objtype),
		size:       size,
	}

	copy(g.id[:], oid[:])
	g.content = make([]byte, size)
	zr.Read(g.content)

	if g.size != int64(len(g.content)) {
		log.Fatal("data corruptted.")
	}

	return g, nil
}

func (g *GitObject) Save(w io.Writer) error {
	wt := zlib.NewWriter(w)
	defer wt.Close()

	// write header
	header := fmt.Sprintf("%s %d\x00", strings.ToLower(g.objectType.String()), g.size)
	_, err := wt.Write([]byte(header))
	_, err = wt.Write(g.content)

	err = wt.Flush()

	return err
}

func (g *GitObject) Id() Hash {
	return g.id
}
func (g *GitObject) Type() ObjectType {
	return g.objectType
}

func (g *GitObject) Size() int64 {
	return g.size
}

func (g *GitObject) Content() []byte {
	return g.content
}

// HashObject read data from a reader and create a GitObject classified by the ObjectType argument.
func HashObject(content []byte, t ObjectType) (*GitObject, error) {
	var err error

	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%s %d\u0000%s", t, len(content), content)

	g := &GitObject{}
	h1 := sha1.Sum(b.Bytes())
	copy(g.id[:], h1[:])
	g.objectType = t
	g.size = int64(len(content))
	g.content = make([]byte, len(content))
	copy(g.content, content)

	return g, err
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

		fmt.Fprintf(w, "%08x  % x  % x  %s\n", addr, b[:8], b[8:], string(b))
		addr += 16
	}
}
