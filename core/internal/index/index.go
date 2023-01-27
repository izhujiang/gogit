package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// .git/index

// part1: Header 12 bytes

// part2: IndexEntries (A number of sorted index entries )
// Index File - Index Entry
// 4 bytes	32-bit created time in seconds	Number of seconds since Jan. 1, 1970, 00:00:00.
// 4 bytes	32-bit created time - nanosecond component	Nanosecond component of the created time in seconds value.
// 4 bytes	32-bit modified time in seconds	Number of seconds since Jan. 1, 1970, 00:00:00.
// 4 bytes	32-bit modified time - nanosecond component	Nanosecond component of the created time in seconds value.
// 4 bytes	device	Metadata associated with the file -- these originate from file attributes used on the Unix OS.
// 4 bytes	inode
// 4 bytes	mode
// 4 bytes	user id
// 4 bytes	group id
// 4 bytes	file content length	Number of bytes of content in the file.
// 20 bytes	SHA-1	Corresponding blob object's SHA-1 value.
// 2 bytes	Flags	(High to low bits)
//			1 bit: assume-valid/assume-unchanged flag
//			1-bit: extended flag (must be 0 for versions less than 3; if 1 then an additional 2 bytes follow before the path\file name)
//			2-bit: merge stage 12-bit: path\file name length (if less than 0xFFF)
// 2 bytes (version 3 or higher) Flags	(High to low bits)
//			1-bit: future use
//			1-bit: skip-worktree flag (sparse checkout)
//			1-bit: intent-to-add flag (git add -N)
//			13-bit: unused, must be zero
// Variable Length	Path/file name	NUL terminated

// Ref: https://learn.microsoft.com/en-us/archive/msdn-magazine/2017/august/devops-git-internals-architecture-and-index-files

// part3(optional): Extensions

// part4: Hash
// Hash checksum (Hash checksum over the content of the index file before this checksum.)

// All binary numbers are in network byte order.
type Header struct {
	Signature       [4]byte // “DIRC” for "dircache"
	Version         uint32
	NumberOfEntries uint32
}

type FileAttribute struct {
	Ctime_seconds     uint32 // 32-bit ctime seconds, the last time a file's metadata changed
	Ctime_nanoseconds uint32 // 32-bit ctime nanosecond fractions
	Mtime_seconds     uint32 // 32-bit mtime seconds, the last time a file's data changed
	Mtime_nanoseconds uint32 // 32-bit mtime nanosecond fractions
	Dev               uint32 // 32-bit dev (divice)
	Ino               uint32 // 32-bit ino (inode)
	Mode              uint32
	// 32-bit mode, split into (high to low bits)
	// 4-bit object type,
	// 3-bit unused
	//  9-bit unix permission
	Uid      uint32 //
	Gid      uint32
	Filesize uint32   // This is the on-disk size from stat(2), truncated to 32-bit.
	ObjectId [20]byte // Object name for the represented object
	Flags    uint16
	//A 16-bit 'flags' field split into (high to low bits)
	// 1-bit assume-valid flag
	// 1-bit extended flag (must be zero in version 2)
	// 2-bit stage (during merge)
	// 12-bit name length if the length is less than 0xFFF; otherwise 0xFFF // is stored in this field.

	// ExtendedFlag uint16 // only applicable if "extented flag is 1"
	// (Version 3 or later) A 16-bit field, only applicable if the
	// "extended flag" above is 1, split into (high to low bits).
	// 1-bit reserved for future
	// 1-bit skip-worktree flag (used by sparse checkout)
	// 1-bit intent-to-add flag (used by "git add -N")
	// 13-bit unused, must be zero
}

// ExtendedFlag uint16
type FileEntry struct {
	FileAttribute

	ExtendedFlag uint16 // only applicable if "extented flag is 1"
	// 1-bit reserved for future
	// 1-bit skip-worktree flag (used by sparse checkout)
	// 1-bit intent-to-add flag (used by "git add -N")
	// 13-bit unused, must be zeroa
	Path []byte // Entry path name (variable length) relative to top level directory (without leading slash). '/' is used as path separator. The special path components ".", ".." and ".git" (without quotes) are disallowed.  Trailing slash is also disallowed.
	// The exact encoding is undefined, but the '.' and '/' characters are encoded in 7-bit ASCII and the encoding cannot contain a NUL byte (iow, this is a UNIX pathname).
	// (Version 4) In version 4, the entry path name is prefix-compressed relative to the path name for the previous entry (the very first entry is encoded as if the path name for the previous entry is an empty string).  At the beginning of an entry, an integer N in the variable width encoding (the same encoding as the offset is encoded for OFS_DELTA pack entries; see pack-format.txt) is stored, followed by a NUL-terminated string S.  Removing N bytes from the end of the path name for the previous entry, and replacing it with the string S yields the path name for this entry.

	padNum int // 1-8 nul bytes as necessary to pad the entry to a multiple of eight bytes while keeping the name NUL-terminated.
	// (Version 4) In version 4, the padding after the pathname does not exist.

}

type Extension struct {
	signature [4]byte // If the first byte is 'A'..'Z' the extension is optional and can be ignored.
	size      uint32  // 32-bit size of the extension
	data      []byte  // Extension data
}

type Index struct {
	Header
	Entries    []*FileEntry
	Extensions []*Extension
	Checksum   [20]byte
}

func Load(dr io.Reader) *Index {
	r := bufio.NewReader(dr)
	var header Header
	binary.Read(r, binary.BigEndian, &header)
	if header.Version != 2 {
		msg := fmt.Sprintf("version %d is not supported yet.", header.Version)
		panic(msg)

	}

	entries := make([]*FileEntry, header.NumberOfEntries)
	var i uint32
	for i = 0; i < header.NumberOfEntries; i++ {
		fileEntry := &FileEntry{}
		fileAttr := &(fileEntry.FileAttribute)
		binary.Read(r, binary.BigEndian, fileAttr)
		fileAttrBytesSize := 62

		// TODO: Read extended flag if version >=3
		// (Version 3 or later) "extended flag" is 1, split into (high to low bits).
		if fileAttr.Flags&0x8000 != 0 {
			binary.Read(r, binary.BigEndian, &(fileEntry.ExtendedFlag))
			fileAttrBytesSize += 2
		}

		filename, _ := r.ReadBytes(0x00)
		lenOfFilename := len(filename)
		fileEntry.Path = make([]byte, lenOfFilename-1)
		copy(fileEntry.Path, filename)
		fmt.Println("file name: ")
		// fmt.Println(string(fileEntry.Path))

		overflow := (fileAttrBytesSize + lenOfFilename) % 8

		if overflow == 0 {
			fileEntry.padNum = 1
		} else {
			skip := 8 - overflow
			r.Discard(skip)
			fileEntry.padNum = skip + 1
		}

		entries[i] = fileEntry
	}

	extensionSig, _ := r.Peek(4)
	if bytes.Equal(extensionSig, []byte("TREE")) {
		panic("Cache tree Extension in index file not implemented")
	} else if bytes.Equal(extensionSig, []byte("REUC")) {
		panic("Resolve undo Extension in index file not implemented")
	} else if bytes.Equal(extensionSig, []byte("link")) {
		panic("Split index Extension in index file not implemented")
	} else if bytes.Equal(extensionSig, []byte("UNTR")) {
		panic("Untracked cache Extension in index file not implemented")
	} else if bytes.Equal(extensionSig, []byte("FSMN")) {
		panic("File System Monitor cache Extension in index file not implemented")
	} else {
		// If the first byte is 'A'..'Z' the extension is optional and can be ignored.
		// TODO: how to distinct optional extension from no extension?
	}
	var h [20]byte
	r.Read(h[:])

	return &Index{
		Header:   header,
		Entries:  entries,
		Checksum: h,
	}

}

func Save(w io.Writer) {
	panic("Save Index")

}
