package ipfs_api

import (
	"log"

	"github.com/ipfs/go-ipfs-api"
)

var sh *shell.Shell

var Addr = "127.0.0.1:5001"

func reset() {
	if sh == nil || !sh.IsUp() {
		sh = shell.NewShell(Addr)
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
