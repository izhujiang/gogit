package porcelain

import (
	"log"

	"github.com/izhujiang/gogit/core"
)

// Init
func Init(root string) error {
	err := core.InitRepository(root)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
