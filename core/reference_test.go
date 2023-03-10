package core

import (
	"testing"
)

func TestHead(t *testing.T) {
	refs := GetReferencs()
	if refs.Head() != "main" {
		t.Error("head of refs is not main")

	}
}
