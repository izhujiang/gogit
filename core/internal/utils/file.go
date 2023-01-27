package utils

import (
	"bytes"
	"io"
	"os"
)

func LoadFile(path string) ([]byte, error) {
	_, err := os.Stat(path)
	var buf bytes.Buffer
	if err != nil {
		return buf.Bytes(), err
	}
	f, _ := os.Open(path)
	defer f.Close()

	_, err = io.Copy(&buf, f)

	return buf.Bytes(), err
}
