package core

import (
	"fmt"
	"testing"

	"github.com/izhujiang/gogit/common"
	"github.com/stretchr/testify/assert"
)

func TestNewObjectId(t *testing.T) {
	s := "8b80381e99f222fb1ffe69a925f5b10ceace5165"
	get, _ := common.NewHash(s)
	want := [20]byte{0x8b, 0x80, 0x38, 0x1e, 0x99, 0xf2, 0x22, 0xfb, 0x1f, 0xfe, 0x69, 0xa9, 0x25, 0xf5, 0xb1, 0x0c, 0xea, 0xce, 0x51, 0x65}
	assert.Equal(t, want, [20]byte(get), fmt.Sprintf("string %s is hashed as:\n% x\nwhich should be:\n% x\n", s, [20]byte(get), want))
}

func TestObjectIdString(t *testing.T) {
	s := "8b80381e99f222fb1ffe69a925f5b10ceace5165"
	h, _ := common.NewHash(s)
	get := h.String()
	want := s
	assert.Equal(t, want, get, fmt.Sprintf("%s should be equal to %s", s, h.String()))

}
