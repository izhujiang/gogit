package object

import (
	"bytes"
	"io"
	"strings"

	"github.com/izhujiang/gogit/common"
)

// GitObject <==> Blob
func GitObjectToBlob(g *GitObject) *Blob {
	// copy(b.oid[:], g.oid[:])
	b := &Blob{
		oid:     g.Hash(),
		content: make([]byte, g.Size()),
	}
	copy(b.content, g.Content())

	return b
}

func BlobToGitObject(b *Blob) *GitObject {
	g := &GitObject{
		objectType: ObjectTypeBlob,
		size:       int64(len(b.content)),
	}

	copy(b.content, g.content)

	return g
}

// GitObject <==> Tree
func GitObjectToTree(g *GitObject) *Tree {
	r := bytes.NewBuffer(g.content)
	entries := newTreeEntryCollecion()

	for {
		mode, err := r.ReadString(0x20)
		mode = strings.Trim(mode, " ")
		if err == io.EOF {
			break
		}
		name, _ := r.ReadBytes(0x00)
		fileName := string(name[:len(name)-1])
		var oid common.Hash
		_, _ = r.Read(oid[:])

		fm, _ := common.NewFileMode(mode)
		entry := NewTreeEntry(oid, fileName, fm)
		entries.add(entry)
	}

	t := &Tree{
		oid: g.Hash(),
	}
	t.entries = entries

	return t
}

func TreeToGitObject(t *Tree) *GitObject {
	w := &bytes.Buffer{}

	entries := t.entries.sort()
	for _, entry := range entries {
		mode := strings.TrimLeft(entry.Mode.String(), "0 ")
		w.WriteString(mode)
		w.WriteByte(0x20)
		w.WriteString(entry.Name)
		w.WriteByte(0x00)
		w.Write(entry.Oid[:])
	}

	content := w.Bytes()

	g := &GitObject{
		objectType: ObjectTypeTree,
		size:       int64(len(content)),
		content:    content,
	}
	return g
}
