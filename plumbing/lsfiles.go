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
	sa := core.GetStagingArea()
	sa.Load()

	if option.Cached {
		sa.ListIndex(w, false)
	}

	if option.Stage {
		sa.ListIndex(w, true)
	}

	return nil
}
