package index

import (
	"bytes"
	"io"
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

	// // TODO: Serialize extentions
	// buf.Write(idx.Extensions)

	// // append checksum
	// h := sha1.Sum(buf.Bytes())
	// buf.Write(h[:])

	// buf.WriteTo(w)
}

func encodeHeader(w io.Writer, idx *Index) {
	Write(w, sign_Index)
	Write(w, idx.Version)
	Write(w, len(idx.Entries))
}

func encodeIndexEntries(w io.Writer, idx *Index) {
	var c_sec, c_nsec uint32
	var m_sec, m_nsec uint32
	var flags uint16

	for _, entry := range idx.Entries {
		c_sec, c_nsec, _ = timeToUint32(entry.CTime)
		m_sec, m_nsec, _ = timeToUint32(entry.MTime)

		Write(w, c_sec)
		Write(w, c_nsec)
		Write(w, m_sec)
		Write(w, m_nsec)

		Write(w, entry.Dev)
		Write(w, entry.Ino)
		Write(w, entry.Mode)

		Write(w, entry.Uid)
		Write(w, entry.Gid)
		Write(w, entry.Size)
		Write(w, entry.Oid[:])

		flags = uint16(entry.Stage&0x3) << 12
		if l := len(entry.Filepath); l < maskFlagNameLength {
			flags |= uint16(l)
		} else {
			flags |= 0x0FFF
		}

		Write(w, flags)

		entry_fixed_size := 62

		if entry.IntentToAdd || entry.Skipworktree {
			var ext_flags uint16
			if entry.IntentToAdd {
				ext_flags |= maskExtflagIntentToAdd
			}
			if entry.Skipworktree {
				ext_flags |= maskExtflagSkipWorktree
			}

			Write(w, ext_flags)
			entry_fixed_size += 2
		}

		Write(w, []byte(entry.Filepath))

		if idx.Version == idx_version_2 || idx.Version == idx_version_3 {
			entrySize := entry_fixed_size + len(entry.Filepath)
			padLen := 8 - entrySize%8
			pad := make([]byte, padLen)
			Write(w, pad)
		} else {
			// do nothing pad
		}
	}
}

func encodeExtensions(w io.Writer, idx *Index) {

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