package plumbing

import (
	"io"

	"github.com/izhujiang/gogit/core"
)

type LsFilesOption struct {
	Cached    bool
	Deleted   bool
	Modified  bool
	Others    bool
	Ignored   bool
	Stage     bool
	Directory bool
	Unmerged  bool
	Killed    bool
}

func LsFiles(w io.Writer, option *LsFilesOption) error {
	// Show cached files in the output (default)
	if option.Cached {
		sa := core.GetStagingArea()
		sa.LsFiles(w, false)
	}

	if option.Stage {
		sa := core.GetStagingArea()
		sa.LsFiles(w, true)
	}

	return nil
}
