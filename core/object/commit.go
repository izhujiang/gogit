package object

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/izhujiang/gogit/common"
)

type Commit struct {
	GitObject
	tree      common.Hash
	parents   []common.Hash
	author    string
	committer string
	message   string
}

func EmptyCommit() *Commit {
	return &Commit{}
}

func NewCommit(oid common.Hash, tree common.Hash, parents []common.Hash, author string, committer string, message string) *Commit {
	return &Commit{
		GitObject: GitObject{
			oid: oid,
		},
		tree:      tree,
		parents:   parents,
		author:    author,
		committer: committer,
		message:   message,
	}
}

func (c *Commit) SetId(oid common.Hash) {
	c.oid = oid
}

func (c *Commit) Tree() common.Hash {
	return c.tree
}
func (c *Commit) Parents() []common.Hash {
	return c.parents
}

func (c *Commit) FromGitObject(g *GitObject) {
	c.GitObject = *g
	c.parseContent()
}

func GitObjectToCommit(g *GitObject) *Commit {
	c := &Commit{
		GitObject: *g,
	}
	c.parseContent()

	return c
}

func (c *Commit) Serialize(w io.Writer) error {
	c.composeContent()

	return c.GitObject.Serialize(w)
}

func (c *Commit) Deserialize(r io.Reader) error {
	err := c.GitObject.Deserialize(r)

	if err != nil {
		return err
	}

	c.parseContent()
	return nil
}

func (c *Commit) Hash() common.Hash {
	c.composeContent()

	return c.GitObject.Hash()
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

func (c *Commit) parseContent() {
	buf := bytes.NewBuffer(c.content)

	space := string([]byte{common.SPACE})
	delim := string([]byte{common.DELIM})

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
}

func (c *Commit) composeContent() {
	w := &bytes.Buffer{}

	w.WriteString(Kind_Commit.String())
	w.WriteByte(common.SPACE)
	w.WriteString(c.tree.String())
	w.WriteByte(common.DELIM)

	for _, p := range c.parents {
		w.WriteString("parent")
		w.WriteByte(common.SPACE)
		w.WriteString(p.String())
		w.WriteByte(common.DELIM)
	}

	w.WriteString("author")
	w.WriteByte(common.SPACE)
	w.WriteString(c.author)
	w.WriteByte(common.DELIM)

	w.WriteString("committer")
	w.WriteByte(common.SPACE)
	w.WriteString(c.committer)
	w.WriteByte(common.DELIM)

	w.WriteByte(common.DELIM)
	w.WriteString(c.message)
	w.WriteByte(common.DELIM)

	c.content = w.Bytes()
}
