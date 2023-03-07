package object

import (
	"bytes"

	"github.com/izhujiang/gogit/common"
)

type Tag struct {
	oid       common.Hash
	refObject common.Hash
	refType   ObjectType
	tagger    string
}

func EmptyTag() *Tag {
	return &Tag{}
}

func NewTag(oid common.Hash, refObject common.Hash, refType ObjectType, tagger string) *Tag {
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
	// r := bytes.NewBuffer(g.content)

	// space := string([]byte{0x20})
	// delim := byte(0x0a)

	// for {
	// 	line, err := r.ReadString(delim)
	// 	line = strings.Trim(line, string([]byte{delim}))
	// 	if err == io.EOF {
	// 		break
	// 	}

	// 	if len(line) == 0 {
	// 		continue
	// 	}

	// 	itemName, itemValue, found := strings.Cut(line, space)

	// 	if found {
	// 		switch itemName {
	// 		case "tree":
	// 			c.tree, _ = common.NewHash(itemValue)

	// 		case "parent":
	// 			h, _ := common.NewHash(itemValue)
	// 			c.parents = append(c.parents, h)

	// 		case "author":
	// 			c.author = itemValue

	// 		case "committer":
	// 			c.committer = itemValue

	// 		default: // git comment message
	// 			c.message = line
	// 		}
	// 	} else {
	// 		c.message = c.message + line
	// 	}

}

func (c *Tag) ToGitObject() *GitObject {
	panic("Not implemented")
	// space := byte(0x20)
	// delim := byte(0x0a)

	// w := &bytes.Buffer{}

	// w.WriteString("tree")
	// w.WriteByte(space)
	// w.WriteString(c.tree.String())
	// w.WriteByte(delim)

	// for _, p := range c.parents {
	// 	w.WriteString("parent")
	// 	w.WriteByte(space)
	// 	w.WriteString(p.String())
	// 	w.WriteByte(delim)
	// }

	// w.WriteString("author")
	// w.WriteByte(space)
	// w.WriteString(c.author)
	// w.WriteByte(delim)

	// w.WriteString("committer")
	// w.WriteByte(space)
	// w.WriteString(c.committer)
	// w.WriteByte(delim)

	// w.WriteByte(delim)
	// w.WriteString(c.message)
	// w.WriteByte(delim)

	// content := w.Bytes()
	// g := &GitObject{
	// 	objectType: ObjectTypeCommit,
	// 	size:       int64(len(content)),
	// 	content:    content,
	// }
	// return g
	return nil
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
