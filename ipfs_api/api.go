package ipfs_api

import (
	"log"

	ipfs "github.com/ipfs/go-ipfs-api"
)

var sh *ipfs.Shell

func reset() {
	if sh == nil || !sh.IsUp() {
		sh = ipfs.NewShell("127.0.0.1:5001")
		if id, err := sh.ID(); err == nil {
			v0, _, _ := sh.Version()
			log.Println("ipfs version", v0, "node", id.ID, "online")
		}
	}
}

func Shell() *ipfs.Shell {
	reset()
	return sh
}

func init() {
	reset()
}
