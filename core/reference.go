package core

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/izhujiang/gogit/common"
)

const (
	ref_head string = "refs/heads/main"
)

type References struct {
	root     string
	headpath string
}

func (r *References) Head() string {
	hp := r.activeHead()
	return filepath.Base(hp)
}
func (r *References) activeHead() string {
	f, err := os.Open(r.headpath)
	if err != nil {
		return ref_head
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	head := scanner.Text()

	_, head, found := strings.Cut(head, ":")
	if found {
		return filepath.Join(r.root, strings.Trim(head, " "))
	} else {
		return filepath.Join(r.root, head)
	}
}

// func (r *References) headpath() {

// }
func (r *References) LastCommit() (common.Hash, error) {
	head := r.activeHead()
	f, err := os.Open(head)
	if err != nil {
		return common.ZeroHash, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	lastCommitText := scanner.Text()
	lastCommitId, err := common.NewHash(lastCommitText)
	if err != nil {
		return common.ZeroHash, err
	}
	return lastCommitId, nil
}

func (r *References) SaveCommit(id common.Hash) error {
	head := r.activeHead()

	f, err := os.OpenFile(head, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(id.String())
	return err
}
