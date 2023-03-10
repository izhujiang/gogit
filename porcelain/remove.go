package porcelain

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/izhujiang/gogit/core"
)

func Remove(paths []string) error {
	expandedPaths := make([]string, 0, 64)
	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}

		if fi.IsDir() {
			filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if d.Type().IsRegular() {
					expandedPaths = append(expandedPaths, path)
				}
				return nil
			})
		} else {
			expandedPaths = append(expandedPaths, path)
		}
	}

	// TODO: check if some files have local modifications

	// if files in working erea == files in index && files == repository

	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}

		if fi.IsDir() {
			os.RemoveAll(path)
		} else {
			os.Remove(path)
		}
	}

	sa := core.GetStagingArea()
	sa.Unstage(expandedPaths)

	return nil
}
