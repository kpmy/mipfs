package ipfs_api

import (
	"log"

	"github.com/ipfs/go-ipfs-api"
)

var sh *shell.Shell

func reset() {
	if sh == nil || !sh.IsUp() {
		sh = shell.NewShell("127.0.0.1:5001")
		if id, err := sh.ID(); err == nil {
			v0, _, _ := sh.Version()
			log.Println("ipfs version", v0, "node", id.ID, "online")
		}
	}
}

func Shell() *shell.Shell {
	reset()
	return sh
}

func init() {
	reset()
}
