package porcelain

import (
	"github.com/izhujiang/gogit/core"
)

type RemoveOption struct {
	Recursive bool
}

func Remove(paths []string, option *RemoveOption) error {
	// TODO: check if some files have local modifications, remove all files and directories of working area if with --force

	// fi, err := os.Stat(path)
	// if fi.IsDir() {
	// 	os.RemoveAll(path)
	// } else {
	// 	os.Remove(path)
	// }

	sa := core.GetStagingArea()
	sa.Load()
	sa.Unstage(paths, option.Recursive)
	err := sa.Save()

	return err
}
