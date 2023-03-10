package index

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/izhujiang/gogit/common"
)

// ReadUint64 reads 8 bytes and returns them as a BigEndian uint32
func ReadUint64(r io.Reader) (uint64, error) {
	var v uint64
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}

	return v, nil
}

// ReadUint32 reads 4 bytes and returns them as a BigEndian uint32
func ReadUint32(r io.Reader) (uint32, error) {
	var v uint32
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}

	return v, nil
}

// ReadUint16 reads 2 bytes and returns them as a BigEndian uint16
func ReadUint16(r io.Reader) (uint16, error) {
	var v uint16
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}

	return v, nil
}

func ReadHash(r io.Reader) (common.Hash, error) {
	var v common.Hash
	if err := binary.Read(r, binary.BigEndian, v[:]); err != nil {
		return common.ZeroHash, err
	}

	return v, nil
}

func ReadSlice(r io.Reader, n int) ([]byte, error) {
	v := make([]byte, n)
	if err := binary.Read(r, binary.BigEndian, v); err != nil {
		return nil, err
	}

	return v, nil

}

func ReadUntil(r *bufio.Reader, delim byte) ([]byte, error) {
	buf, err := r.ReadBytes(delim)
	if len(buf) == 0 {
		return buf, err
	} else {
		return buf[:len(buf)-1], err
	}
}

// Write writes the binary representation of data into w. Data must be a fixed-size value or a slice of fixed-size values, or a pointer to such data. Boolean values encode as one byte: 1 for true, and 0 for false.
// Bytes written to w are encoded using the specified byte order (BigEndian) and read from successive fields of the data. When writing structs, zero values are written for fields with blank (_) field names.
func Write[T any](w io.Writer, v T) error {
	if err := binary.Write(w, binary.BigEndian, v); err != nil {
		return err
	}

	return nil
}

func WriteString(w io.Writer, s string) error {
	v := []byte(s)
	if err := binary.Write(w, binary.BigEndian, v); err != nil {
		return err
	}

	return nil
}
