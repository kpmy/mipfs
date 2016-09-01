package ipfs_api

import (
	"log"

	go_ipfs_api "github.com/ipfs/go-ipfs-api"
)

var sh *go_ipfs_api.Shell

func reset() {
	if sh == nil || !sh.IsUp() {
		sh = go_ipfs_api.NewShell("127.0.0.1:5001")
		if id, err := sh.ID(); err == nil {
			v0, _, _ := sh.Version()
			log.Println("ipfs version", v0, "node", id.ID, "online")
		}
	}
}

func Shell() *go_ipfs_api.Shell {
	reset()
	return sh
}

func init() {
	reset()
}
