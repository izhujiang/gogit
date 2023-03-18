package plumbing

import (
	"bytes"
	"testing"

	"github.com/izhujiang/gogit/common"
	"github.com/izhujiang/gogit/core/object"
)

func TestFunction(t *testing.T) {
	input := "test content\n"
	buf := bytes.NewBufferString(input)
	option := &HashObjectOption{
		ObjectType: object.Kind_Blob,
		Write:      false,
	}

	got, _ := HashObject(buf, option)
	expect, _ := common.NewHash("d670460b4b4aece5915caf5c68d12f560a9fe3e4")
	if got != expect {
		t.Fatal("Hash object expect ", expect, "got ", got)
	}
}
