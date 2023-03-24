package index

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"github.com/izhujiang/gogit/common"
)

type IndexDecoder struct {
	Reader io.Reader
}

func NewIndexDecoder(r io.Reader) *IndexDecoder {
	return &IndexDecoder{
		Reader: r,
	}

}

func (id *IndexDecoder) Decode(idx *Index) error {
	// load all contents
	buf := &bytes.Buffer{}
	io.Copy(buf, id.Reader)

	if err := validateCheckSum(buf); err != nil {
		return err
	}

	r := bufio.NewReader(io.LimitReader(buf, int64(buf.Len()-20)))

	err := decodeHeader(r, idx)
	if err != nil {
		return err
	}

	err = decodeIndexEntries(r, idx)
	if err != nil {
		return err
	}

	err = decodeExtensions(r, idx)

	return err
}

func decodeHeader(r io.Reader, idx *Index) error {
	sign, err := ReadSlice(r, 4)
	if err != nil || !bytes.Equal(sign, []byte(sign_Index)) {
		return ErrNotOrInvalidIndexFile
	}

	idx.version, err = ReadUint32(r)
	if err != nil || idx.version != idx_version_2 {
		return ErrInvalidIndexFileVersion
	}

	idx.numberOfIndexEntries, err = ReadUint32(r)
	return err
}

func validateCheckSum(buf *bytes.Buffer) error {
	data := buf.Bytes()
	n := len(data)
	h1 := data[n-20 : n]
	h2 := sha1.Sum(data[:n-20])
	if !bytes.Equal(h1[:], h2[:]) {
		return ErrCorruptedIndexFile
	}

	return nil
}

func decodeIndexEntries(r *bufio.Reader, idx *Index) error {
	// entries := make([]*IndexEntry, 0)
	var c_sec, c_nsec uint32
	var m_sec, m_nsec uint32
	var dev, ino uint32
	var mode uint32
	var uid, gid uint32
	var size uint32
	var flags uint16
	var ext_flags uint16
	var fpath []byte
	var fpathLength int

	for i := 0; i < int(idx.numberOfIndexEntries); i++ {
		c_sec, _ = ReadUint32(r)
		c_nsec, _ = ReadUint32(r)
		m_sec, _ = ReadUint32(r)
		m_nsec, _ = ReadUint32(r)
		dev, _ = ReadUint32(r)
		ino, _ = ReadUint32(r)
		mode, _ = ReadUint32(r)
		uid, _ = ReadUint32(r)
		gid, _ = ReadUint32(r)
		size, _ = ReadUint32(r)
		oid, _ := ReadHash(r)
		flags, _ = ReadUint16(r)

		// version validation
		if (idx.version == idx_version_2) && (flags&maskFlagEntryExtended != 0) {
			fmt.Println(ErrNotOrInvalidIndexFile)
			return ErrNotOrInvalidIndexFile
		}

		// Parse flag
		// (Version 3 or later) "extended flag" is 1, split into (high to low bits).
		entry_fixed_size := 62
		if flags&maskFlagEntryExtended != 0 {
			ext_flags, _ = ReadUint16(r)
			entry_fixed_size += 2

			// 13-bit unused, must be zero
			if ext_flags&maskExtflagUnsed != 0 {
				return ErrNotOrInvalidIndexFile
			}
		}

		fpathLength = int(flags & maskFlagNameLength)

		// read path
		if idx.version == idx_version_2 || idx.version == idx_version_3 {
			// ReadBytes include 0x00
			fpath, _ = ReadUntil(r, sep_NULL)
			// TODO: compare the length of filename with filenameLength
			// OR read filenameLength bytes from io.Reader
			overflow := (entry_fixed_size + len(fpath) + 1) % 8
			if overflow != 0 {
				skip := 8 - overflow
				r.Discard(skip)
			}
		} else { // idx_version_4
			fpath, _ = ReadUntil(r, sep_NULL)
		}

		// validate filepath  length
		if (fpathLength < maskFlagNameLength && fpathLength != len(fpath)) ||
			(fpathLength == maskFlagNameLength && len(fpath) < maskFlagNameLength) {
			return ErrNotOrInvalidIndexFile
		}

		entry := &IndexEntry{
			oid:      oid,
			filepath: string(fpath),
			fileinfo: fileinfo{
				name:  filepath.Base(string(fpath)),
				cTime: time.Unix(int64(c_sec), int64(c_nsec)),
				mTime: time.Unix(int64(m_sec), int64(m_nsec)),
				dev:   dev,
				ino:   ino,
				mode:  common.FileMode(mode),
				uid:   uid,
				gid:   gid,
				size:  size,
			},
			stage:        Stage(flags >> 12 & 0x3),
			skipworktree: (ext_flags & maskExtflagSkipWorktree) != 0,
			intentToAdd:  (ext_flags & maskExtflagIntentToAdd) != 0,
		}
		// fmt.Println("file name: ", string(entry.Filepath))

		idx.entries = append(idx.entries, entry)
	}
	// fmt.Println("idx.entries:", idx.entries)
	return nil
}

func decodeExtensions(r *bufio.Reader, idx *Index) error {
	idx.unknownExtensions = make([]*Extension, 0)
	for {
		sign, err := r.Peek(4)

		if err != nil {
			return err
		}

		switch {
		case bytes.Equal(sign, []byte(sign_ext_Tree)):
			decodeTreeCacheExtension(r, idx)
			// TODO: other extensions
		// case bytes.Equal(sign, []byte(sign_ext_ResolveUndo)):
		default:
			// If the first byte is 'A'..'Z' the extension is optional and can be ignored.
			// if extensionSig[0] >= 0x41 && extensionSig[0] <= 0x5A {
			// unknown extension, just save temporally
			fmt.Println("Unknow extention:", string(sign))
			sign, _ = ReadSlice(r, 4)
			size, _ := ReadUint32(r)
			data, _ := ReadSlice(r, int(size))
			ext := &Extension{
				Signature: sign,
				Size:      size,
				Data:      data,
			}
			idx.unknownExtensions = append(idx.unknownExtensions, ext)
		}

	}
}

func decodeTreeCacheExtension(rd io.Reader, idx *Index) error {
	sign, _ := ReadSlice(rd, 4)
	if !bytes.Equal(sign, []byte(sign_ext_Tree)) {
		msg := "Invalid Signature: " + string(sign)
		panic(msg)

	}
	size, _ := ReadUint32(rd)

	r := bufio.NewReader(io.LimitReader(rd, int64(size)))

	idx.cacheTree = newCacheTree()
	for {
		name, err := ReadUntil(r, sep_NULL)
		if err != nil {
			break
		}
		s_entry_count, _ := ReadUntil(r, sep_SPACE)
		s_subtrees_count, _ := ReadUntil(r, sep_NEWLINE)
		entry_count, _ := strconv.Atoi(string(s_entry_count))
		subtrees_count, _ := strconv.Atoi(string(s_subtrees_count))

		// only the TreeEntry is valid
		var oid []byte
		if entry_count >= 0 {
			oid, _ = ReadSlice(r, 20)
		} else {
			oid = make([]byte, 20)
		}

		te := &CacheTreeEntry{
			Name:         string(name),
			EntryCount:   entry_count,
			SubtreeCount: subtrees_count,
		}
		copy(te.Oid[:], oid)

		idx.cacheTree.cacheTreeEntries = append(idx.cacheTree.cacheTreeEntries, te)
	}
	return nil
}
