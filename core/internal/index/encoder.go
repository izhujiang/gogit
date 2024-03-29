package index

import (
	"bytes"
	"crypto/sha1"
	"io"
	"strconv"
	"time"
)

type IndexEncoder struct {
	Writer io.Writer
	buf    *bytes.Buffer
}

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

func NewIndexEncoder(w io.Writer) *IndexEncoder {
	buf := &bytes.Buffer{}
	return &IndexEncoder{
		Writer: io.MultiWriter(w, buf),
		buf:    buf,
	}
}
func (ie *IndexEncoder) Encode(idx *Index) {
	encodeHeader(ie.Writer, idx)
	encodeIndexEntries(ie.Writer, idx)

	encodeExtensions(ie.Writer, idx)
	encodeChecksum(ie.Writer, ie.buf)

}

func encodeHeader(w io.Writer, idx *Index) {
	buf := &bytes.Buffer{}
	WriteString(buf, sign_Index)
	Write(buf, idx.version)
	Write(buf, uint32(len(idx.IndexEntries.entries)))
	Write(w, buf.Bytes())
}

func encodeIndexEntries(w io.Writer, idx *Index) {
	encodeIndexEntry := func(e *IndexEntry) {
		c_sec, c_nsec, _ := timeToUint32(e.cTime)
		m_sec, m_nsec, _ := timeToUint32(e.mTime)

		Write(w, uint32(c_sec))
		Write(w, uint32(c_nsec))
		Write(w, uint32(m_sec))
		Write(w, uint32(m_nsec))

		Write(w, uint32(e.dev))
		Write(w, uint32(e.ino))
		Write(w, uint32(e.mode))

		Write(w, uint32(e.uid))
		Write(w, uint32(e.gid))
		Write(w, uint32(e.size))
		Write(w, e.oid[:])

		var flags uint16
		flags = uint16(e.stage&0x3) << 12
		if l := len(e.filepath); l < maskFlagNameLength {
			flags |= uint16(l)
		} else {
			flags |= 0x0FFF
		}

		Write(w, uint16(flags))

		entry_fixed_size := 62

		if e.intentToAdd || e.skipworktree {
			var ext_flags uint16
			if e.intentToAdd {
				ext_flags |= maskExtflagIntentToAdd
			}
			if e.skipworktree {
				ext_flags |= maskExtflagSkipWorktree
			}

			Write(w, ext_flags)
			entry_fixed_size += 2
		}

		WriteString(w, e.filepath)

		if idx.version == idx_version_2 || idx.version == idx_version_3 {
			entrySize := entry_fixed_size + len(e.filepath)
			padLen := 8 - entrySize%8
			pad := make([]byte, padLen)
			Write(w, pad)
		} else {
			// do nothing pad
		}
	}

	idx.Foreach(encodeIndexEntry)
}

func encodeExtensions(w io.Writer, idx *Index) {
	encodeExtensionTreeCache(w, idx)

	// TODO: encode other extentions
	for _, ext := range idx.unknownExtensions {
		Write(w, ext.Signature)
		Write(w, ext.Size)
		Write(w, ext.Data)
	}

}

// assuming cacheTree is not nil
func encodeExtensionTreeCache(w io.Writer, idx *Index) {
	if len(idx.CacheTree.cacheTreeEntries) > 0 {
		var size uint32
		data := &bytes.Buffer{}

		idx.CacheTree.foreach(func(e *CacheTreeEntry) {
			// fmt.Println("entry name: ", e.Oid, e.Name, strconv.Itoa(e.EntryCount), strconv.Itoa(e.SubtreeCount))
			WriteString(data, e.Name)
			Write(data, sep_NULL)
			WriteString(data, strconv.Itoa(e.EntryCount))
			Write(data, sep_SPACE)
			WriteString(data, strconv.Itoa(e.SubtreeCount))
			Write(data, sep_NEWLINE)
			if e.EntryCount >= 0 {
				Write(data, e.Oid[:])
			}
		})
		Write(w, []byte(sign_ext_Tree))
		size = uint32(data.Len())
		Write(w, size)
		Write(w, data.Bytes())
	}
}

func encodeChecksum(w io.Writer, buf *bytes.Buffer) {
	h := sha1.Sum(buf.Bytes())
	w.Write(h[:])
	// fmt.Println("sumcheck:", common.Hash(h).String())
}

func timeToUint32(t time.Time) (uint32, uint32, error) {
	if t.IsZero() {
		return 0, 0, nil
	}

	if t.Unix() < 0 || t.UnixNano() < 0 {
		return 0, 0, ErrInvalidTimestamp
	}

	return uint32(t.Unix()), uint32(t.UnixNano()), nil
}
