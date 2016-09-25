package ipfs_api

import (
	"log"

	"github.com/ipfs/go-ipfs-api"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multiaddr-net"
	"github.com/streamrail/concurrent-map"
	"net/http"
)

var sh *MyShell

var Addr = "127.0.0.1:5001"

type MyShell struct {
	shell.Shell
	Url    string
	Client *http.Client
	cache  cmap.ConcurrentMap
}

func reset() {
	if sh == nil || !sh.IsUp() {
		u := Addr
		if a, err := multiaddr.NewMultiaddr(u); err == nil {
			_, host, err := manet.DialArgs(a)
			if err == nil {
				u = host
			}
		}
		sh = &MyShell{
			Url:    u,
			Client: http.DefaultClient,
			cache:  cmap.New(),
		}
		sh.Shell = *shell.NewShellWithClient(sh.Url, sh.Client)
		if id, err := sh.ID(); err == nil {
			v0, _, _ := sh.Version()
			log.Println("ipfs version", v0, "node", id.ID, "online")
		} else {
			log.Fatal("ipfs", err)
			sh = nil
		}
	}
}

func Shell() *MyShell {
	reset()
	return sh
}

func init() {
	reset()
}
