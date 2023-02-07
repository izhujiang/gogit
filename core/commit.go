package core

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/izhujiang/gogit/common"
)

type Commit struct {
	oid       common.Hash
	tree      common.Hash
	parent    []common.Hash
	author    string
	committer string
	message   string
}

func GitObjectToCommit(g *GitObject) *Commit {
	commit := &Commit{}
	// copy(commit.oid[:], g.oid[:])

	commit.decodeContent(g.content)
	return commit
}

func CommitToGitObject(c *Commit) *GitObject {
	panic("CommitToGitObject not implemented")
}

func (c *Commit) decodeContent(content []byte) {
	r := bytes.NewBuffer(content)

	space := string([]byte{0x20})
	delim := byte(0x0a)

	for {
		line, err := r.ReadString(delim)
		line = strings.Trim(line, string([]byte{delim}))
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
				c.parent = append(c.parent, h)

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

// TODO: output with format interface
func (c *Commit) ShowContent(w io.Writer) {
	fmt.Fprintf(w, "%s %s \n", "tree", c.tree)
	for _, p := range c.parent {
		fmt.Fprintf(w, "%s %s \n", "parent", p)
	}

	fmt.Fprintf(w, "%s %s \n", "author", c.author)
	fmt.Fprintf(w, "%s %s \n", "committer", c.committer)

	fmt.Println("")
	fmt.Println("")
	fmt.Fprintf(w, "%s\n", c.message)
}
