package object

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

	t.decodeContent(g.content)
	return t
}

func (t *Tag) decodeContent(content []byte) {
	panic("Not implemented")

}
