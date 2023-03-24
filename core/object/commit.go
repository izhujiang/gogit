package object

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/izhujiang/gogit/common"
)

type TreePair struct {
	tree        common.Hash
	treePointer *Tree
}
type Commit struct {
	oid common.Hash
	TreePair
	parents   []common.Hash
	author    string
	committer string
	message   string
}

func (c *Commit) Id() common.Hash {
	return c.oid
}

func (c *Commit) Kind() ObjectKind {
	return Kind_Commit
}

func EmptyCommit() *Commit {
	return &Commit{}
}

func NewCommit(oid common.Hash, treeId common.Hash, parents []common.Hash, author string, committer string, message string) *Commit {
	return &Commit{
		oid: oid,
		TreePair: TreePair{
			tree: treeId,
		},
		parents:   parents,
		author:    author,
		committer: committer,
		message:   message,
	}
}

func (c *Commit) Tree() common.Hash {
	return c.tree
}
func (c *Commit) Parents() []common.Hash {
	return c.parents
}

// GitObject ==> Tree,fitll Tree using GotObject from repository
func GitObjectToCommit(g *GitObject) *Commit {
	buf := bytes.NewBuffer(g.content)

	space := string([]byte{common.SPACE})
	delim := string([]byte{common.DELIM})

	c := EmptyCommit()
	c.oid = g.oid
	for {
		line, err := buf.ReadString(common.DELIM)
		line = strings.Trim(line, delim)
		if err == io.EOF {
			break
		}

		if len(line) == 0 {
			continue
		}

		itemName, itemValue, found := strings.Cut(line, space)

		if found {
			switch itemName {
			case "tree":
				c.tree, _ = common.NewHash(itemValue)

			case "parent":
				h, _ := common.NewHash(itemValue)
				c.parents = append(c.parents, h)

			case "author":
				c.author = itemValue

			case "committer":
				c.committer = itemValue

			default: // git comment message
				c.message = line
			}
		} else {
			c.message = c.message + line
		}
	}

	return c
}

func (c *Commit) ToGitObject() *GitObject {
	content := c.contentToBytes()
	g := NewGitObject(Kind_Commit, content)

	return g
}

func (c *Commit) Hash() common.Hash {
	content := c.contentToBytes()
	c.oid = common.HashObject(c.Kind().String(), content)
	return c.oid

}

func (c *Commit) contentToBytes() []byte {
	buf := &bytes.Buffer{}

	buf.WriteString(Kind_Tree.String())
	buf.WriteByte(common.SPACE)
	buf.WriteString(c.tree.String())
	buf.WriteByte(common.DELIM)

	for _, p := range c.parents {
		buf.WriteString("parent")
		buf.WriteByte(common.SPACE)
		buf.WriteString(p.String())
		buf.WriteByte(common.DELIM)
	}

	buf.WriteString("author")
	buf.WriteByte(common.SPACE)
	buf.WriteString(c.author)
	buf.WriteByte(common.DELIM)

	buf.WriteString("committer")
	buf.WriteByte(common.SPACE)
	buf.WriteString(c.committer)
	buf.WriteByte(common.DELIM)

	buf.WriteByte(common.DELIM)
	buf.WriteString(c.message)
	buf.WriteByte(common.DELIM)

	return buf.Bytes()
}

// TODO: output with format interface
func (c *Commit) Content() string {
	w := &bytes.Buffer{}
	fmt.Fprintf(w, "%s %s \n", "tree", c.tree)
	for _, p := range c.parents {
		fmt.Fprintf(w, "%s %s \n", "parent", p)
	}

	fmt.Fprintf(w, "%s %s \n", "author", c.author)
	fmt.Fprintf(w, "%s %s \n", "committer", c.committer)

	fmt.Fprintln(w, "")
	// fmt.Fprintln(w, "")
	fmt.Fprintf(w, "%s\n", c.message)

	return string(w.Bytes())
}
