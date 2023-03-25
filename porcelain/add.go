package porcelain

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/izhujiang/gogit/core"
)

// Add file to repository and add index to stage area
func Add(paths []string) error {
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

	sa := core.GetStagingArea()
	sa.Load()
	sa.Stage(expandedPaths)
	sa.Save()

	return nil
}
