package ipfs_api

import (
	"log"

	"github.com/ipfs/go-ipfs-api"
	"net/http"
)

var sh *MyShell

var Addr = "127.0.0.1:5001"

type MyShell struct {
	shell.Shell
	Url    string
	Client *http.Client
}

func reset() {
	if sh == nil {
		sh = &MyShell{
			Url:    Addr,
			Client: http.DefaultClient,
		}
		sh.Shell = *shell.NewShellWithClient(sh.Url, sh.Client)
		if id, err := sh.ID(); err == nil {
			v0, _, _ := sh.Version()
			log.Println("ipfs version", v0, "node", id.ID, "online")
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
