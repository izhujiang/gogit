package core

type Tag struct {
	gObj          *GitObject
	refObject     string // ref to an object
	refObjectType string // commit
	tagger        string
	message       string
}

func NewTag(g *GitObject) *Tag {
	t := &Tag{
		gObj: g,
	}

	t.parseContent(g.content)
	return t
}

func (t *Tag) parseContent(content []byte) {
	panic("Not implemented")

}
