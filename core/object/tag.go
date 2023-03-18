package object

import (
	"bytes"

	"github.com/izhujiang/gogit/common"
)

type Tag struct {
	oid       common.Hash
	refObject common.Hash
	refType   ObjectKind
	tagger    string
}

func EmptyTag() *Tag {
	return &Tag{}
}

func NewTag(oid common.Hash, refObject common.Hash, refType ObjectKind, tagger string) *Tag {
	return &Tag{
		oid,
		refObject,
		refType,
		tagger,
	}
}
func (c *Tag) Id() common.Hash {
	h := c.oid
	return h
}
func (c *Tag) SetId(oid common.Hash) {
	c.oid = oid
}

func (c *Tag) FromGitObject(g *GitObject) {
	panic("Not implemented")
}

func (c *Tag) ToGitObject() *GitObject {
	panic("Not implemented")
}

// TODO: output with format interface
func (c *Tag) Content() string {
	w := &bytes.Buffer{}
	// fmt.Fprintf(w, "%s %s \n", "tree", c.tree)
	// for _, p := range c.parents {
	// 	fmt.Fprintf(w, "%s %s \n", "parent", p)
	// }

	// fmt.Fprintf(w, "%s %s \n", "author", c.author)
	// fmt.Fprintf(w, "%s %s \n", "committer", c.committer)

	// fmt.Fprintln(w, "")
	// // fmt.Fprintln(w, "")
	// fmt.Fprintf(w, "%s\n", c.message)

	return string(w.Bytes())
}
