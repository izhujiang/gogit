package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sort"
	"time"

	"github.com/izhujiang/gogit/common"
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

type IndexEntryFixedPart struct {
	// 32-bit ctime seconds, the last time a file's metadata changed
	Ctime_seconds uint32
	// 32-bit ctime nanosecond fractions
	Ctime_nanoseconds uint32
	// 32-bit mtime seconds, the last time a file's data changed
	Mtime_seconds uint32
	// 32-bit mtime nanosecond fractions
	Mtime_nanoseconds uint32
	// 32-bit dev (divice)
	Dev uint32
	// 32-bit ino (inode)
	Ino uint32
	// 32-bit mode, split into (high to low bits)
	// 4-bit object type,
	// 3-bit unused
	//  9-bit unix permission
	Mode uint32
	Uid  uint32
	Gid  uint32

	// This is the on-disk size from stat(2), truncated to 32-bit.
	Filesize uint32

	// Object name for the represented object
	ObjectId common.Hash

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
	Flags uint16
}

func (a IndexEntryFixedPart) StageNo() uint16 {
	sn := (a.Flags & 0x0300) >> 8
	return sn
}

// ExtendedFlag uint16
type IndexEntry struct {
	IndexEntryFixedPart

	// only applicable if "extented flag is 1"
	// 1-bit reserved for future
	// 1-bit skip-worktree flag (used by sparse checkout)
	// 1-bit intent-to-add flag (used by "git add -N")
	// 13-bit unused, must be zeroa
	ExtendedFlag uint16

	// Entry path name (variable length) relative to top level directory (without leading slash). '/' is used as path separator. The special path components ".", ".." and ".git" (without quotes) are disallowed.  Trailing slash is also disallowed.
	// The exact encoding is undefined, but the '.' and '/' characters are encoded in 7-bit ASCII and the encoding cannot contain a NUL byte (iow, this is a UNIX pathname).
	// (Version 4) In version 4, the entry path name is prefix-compressed relative to the path name for the previous entry (the very first entry is encoded as if the path name for the previous entry is an empty string).  At the beginning of an entry, an integer N in the variable width encoding (the same encoding as the offset is encoded for OFS_DELTA pack entries; see pack-format.txt) is stored, followed by a NUL-terminated string S.  Removing N bytes from the end of the path name for the previous entry, and replacing it with the string S yields the path name for this entry.
	Path []byte

	// 1-8 nul bytes as necessary to pad the entry to a multiple of eight bytes while keeping the name NUL-terminated.
	// (Version 4) In version 4, the padding after the pathname does not exist.
	padNum int
}

func NewIndexEntry(oid common.Hash, mode uint32, path string) *IndexEntry {
	ie := &IndexEntry{
		IndexEntryFixedPart: IndexEntryFixedPart{
			Mode: mode,
		},
		Path: []byte(path),
	}
	copy(ie.ObjectId[:], oid[:])

	// A 16-bit 'flags' field split into (high to low bits)
	// 1-bit assume-valid flag
	// 1-bit extended flag (must be zero in version 2)
	// 2-bit stage (during merge)
	// 12-bit name length if the length is less than 0xFFF; otherwise 0xFFF is stored in this field.
	lenPath := len(ie.Path)
	if lenPath < 0xFFF {
		ie.Flags = uint16(lenPath)
	} else {
		ie.Flags = 0xFFF
	}

	return ie
}

// TODO: function NewIndexEntry with more details

type ExtensionHeader struct {
	Signature [4]byte // If the first byte is 'A'..'Z' the extension is optional and can be ignored.
	Size      uint32  // 32-bit size of the extension
}
type Extension struct {
	ExtensionHeader
	Data []byte // Extension data
}

type IndexFile struct {
	Header
	Entries    []*IndexEntry
	Extensions []byte
	// Extensions []*Extension
	// Checksum [20]byte
}

func Load(dr io.Reader) *IndexFile {
	r := &bytes.Buffer{}
	n, _ := r.ReadFrom(dr)

	data := r.Bytes()
	h1 := data[n-20 : n]
	h2 := sha1.Sum(data[:n-20])
	if bytes.Compare(h1[:], h2[:]) != 0 {
		log.Fatal("corrupted index file")
	}

	var header Header
	binary.Read(r, binary.BigEndian, &header)
	if header.Version != 2 {
		msg := fmt.Sprintf("version %d is not supported yet.", header.Version)
		panic(msg)
	}

	entries := make([]*IndexEntry, header.NumberOfEntries)
	var i uint32
	for i = 0; i < header.NumberOfEntries; i++ {
		fileEntry := &IndexEntry{}
		fileAttr := &(fileEntry.IndexEntryFixedPart)
		binary.Read(r, binary.BigEndian, fileAttr)
		fileAttrBytesSize := 62

		// TODO: Read extended flag if version >=3
		// (Version 3 or later) "extended flag" is 1, split into (high to low bits).
		if fileAttr.Flags&0x8000 != 0 {
			binary.Read(r, binary.BigEndian, &(fileEntry.ExtendedFlag))
			fileAttrBytesSize += 2
		}

		filename, _ := r.ReadBytes(0x00) // ReadBytes include 0x00
		lenOfFilename := len(filename)
		fileEntry.Path = make([]byte, lenOfFilename-1)
		copy(fileEntry.Path, filename)
		// fmt.Println("file name: ", string(filename))
		// fmt.Println(string(fileEntry.Path), len([]byte(fileEntry.Path)))

		overflow := (fileAttrBytesSize + lenOfFilename) % 8

		if overflow == 0 {
			fileEntry.padNum = 1
		} else {
			skip := 8 - overflow
			// r.Discard(skip)
			empty := make([]byte, skip)
			r.Read(empty)

			fileEntry.padNum = skip + 1
		}

		entries[i] = fileEntry
	}

	// extentions := []*Extension{}
	// for {
	// 	extensionSig, _ := r.Peek(4)

	// 	var extensionHeader ExtensionHeader

	// 	if bytes.Equal(extensionSig, []byte("TREE")) {
	// 		binary.Read(r, binary.BigEndian, &extensionHeader)
	// 		p := make([]byte, extensionHeader.Size)
	// 		r.Read(p)
	// 		ext := &Extension{
	// 			ExtensionHeader: extensionHeader,
	// 			Data:            p,
	// 		}
	// 		extentions = append(extentions, ext)

	// 	} else if bytes.Equal(extensionSig, []byte("REUC")) {
	// 		panic("Resolve undo Extension in index file not implemented")
	// 	} else if bytes.Equal(extensionSig, []byte("link")) {
	// 		panic("Split index Extension in index file not implemented")
	// 	} else if bytes.Equal(extensionSig, []byte("UNTR")) {
	// 		panic("Untracked cache Extension in index file not implemented")
	// 	} else if bytes.Equal(extensionSig, []byte("FSMN")) {
	// 		panic("File System Monitor cache Extension in index file not implemented")
	// 	} else {
	// 		// If the first byte is 'A'..'Z' the extension is optional and can be ignored.
	// 		// TODO: how to distinct optional extension from no extension?
	// 		if extensionSig[0] >= 0x41 && extensionSig[0] <= 0x5A {
	// 			binary.Read(r, binary.BigEndian, &extensionHeader)
	// 			p := make([]byte, extensionHeader.Size)
	// 			r.Read(p)

	// 			// TODO: to deal with optional extention
	// 		} else {
	// 			break
	// 		}
	// 	}
	// }
	// var h [20]byte
	// r.Read(h[:])
	// TODO: read extentions, remember to fix Save() as well

	rest := &bytes.Buffer{}
	r.WriteTo(rest)

	extentions := rest.Bytes()[:rest.Len()-20]

	fmt.Println("extensions:", extentions)

	return &IndexFile{
		Header:     header,
		Entries:    entries,
		Extensions: extentions,
	}

	// return &IndexFile{
	// 	Header:     header,
	// 	Entries:    entries,
	// 	Extensions: extentions,
	// 	Checksum:   h,
	// }

}

func (idx *IndexFile) Save(w io.Writer) {
	// buffering
	buf := &bytes.Buffer{}
	idx.NumberOfEntries = uint32(len(idx.Entries))
	binary.Write(buf, binary.BigEndian, idx.Header)

	for _, entry := range idx.Entries {
		binary.Write(buf, binary.BigEndian, entry.IndexEntryFixedPart)

		fileAttrBytesSize := 62
		if entry.Flags&0x8000 != 0 {
			binary.Write(buf, binary.BigEndian, entry.ExtendedFlag)
			fileAttrBytesSize += 2
		}
		buf.WriteString(string(entry.Path))
		buf.WriteByte(0x00)

		lenOfFilename := len(entry.Path) + 1
		padNum := 8 - (fileAttrBytesSize+lenOfFilename)%8

		if padNum != 8 {
			pad := make([]byte, padNum)
			buf.Write(pad)
		}
	}

	// TODO: Serialize extentions
	buf.Write(idx.Extensions)

	// append checksum
	h := sha1.Sum(buf.Bytes())
	buf.Write(h[:])

	buf.WriteTo(w)
}

func (idx *IndexFile) RemoveAll() {
	idx.Entries = make([]*IndexEntry, 0)
	idx.NumberOfEntries = 0

}
func (idx *IndexFile) InsertEntry(entry *IndexEntry) {
	// add, sort and update header
	idx.Entries = append(idx.Entries, entry)
	idx.NumberOfEntries += 1
	sort.SliceStable(idx.Entries, func(i, j int) bool {
		return bytes.Compare(idx.Entries[i].Path, idx.Entries[j].Path) < 0
	})
}

func (idx *IndexFile) InsertEntries(entries []*IndexEntry) {
	// add, sort and update header
	idx.Entries = append(idx.Entries, entries...)
	// idx.NumberOfEntries += uint32(len(entries))
	sort.SliceStable(idx.Entries, func(i, j int) bool {
		return bytes.Compare(idx.Entries[i].Path, idx.Entries[j].Path) < 0
	})

	idx.NumberOfEntries = uint32(len(idx.Entries))

}

// Dump index file
func (idx *IndexFile) Dump(w io.Writer) {
	headerformat := "%-40s %7s %8s %4s %4s %8s %8s %20s %20s  %-20s\n"
	fmt.Fprintf(w,
		headerformat,
		"Oid",
		"Mode",
		"Size",
		"Uid",
		"Gid",
		"Dev",
		"Ino",
		"Mtime",
		"Ctime",
		"Filepath",
	)

	lineFormat := "%20s %#o %8d %04d %4d %8d %8d %20v %20v  %s\n"
	for _, entry := range idx.Entries {
		// fmt.Fprintf(w, "%o %s %d \t%s\n", entry.Mode, entry.ObjectId, entry.StageNo(), string(entry.Path))
		fmt.Fprintf(w,
			lineFormat,
			entry.ObjectId,
			entry.Mode,
			entry.Filesize,
			entry.Uid,
			entry.Gid,
			entry.Dev,
			entry.Ino,
			time.Unix(int64(entry.Mtime_seconds), int64(entry.Mtime_nanoseconds)).Format("2006-01-02T15:04:05"),
			time.Unix(int64(entry.Ctime_seconds), int64(entry.Mtime_nanoseconds)).Format("2006-01-02T15:04:05"),
			string(entry.Path),
		)
	}
}
