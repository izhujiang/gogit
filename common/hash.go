package common

import (
	"encoding/hex"
	"errors"
)

// Git Object Id represented with 20 bytes
type Hash [20]byte

var InvalidObjectId Hash
var (
	ErrMalformatedString = errors.New("malformed hexadecimal hash represented string")
)

var (
	ZeroHash Hash
)

// NewHash return a new Hash from a hexadecimal hash representation
// "8b80381e99f222fb1ffe69a925f5b10ceace5165" => [ 8b 80 38 1e 99 f2 22 fb 1f fe 69 a9 25 f5 b1 0c ea ce 51 65]
func NewHash(s string) (Hash, error) {
	var h Hash
	if len(s) != 40 {
		return h, ErrMalformatedString
	}

	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 20 {
		return h, ErrMalformatedString
	}
	copy(h[:], b)
	return h, nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}
