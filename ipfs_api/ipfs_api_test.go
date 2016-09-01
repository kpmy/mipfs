package ipfs_api

import (
	"testing"

	"net/http"

	ipfs "github.com/ipfs/go-ipfs-api"
)

func TestShell(t *testing.T) {
	sh := ipfs.NewShellWithClient("127.0.0.1:5001", http.DefaultClient)
	id, _ := sh.ID()
	t.Log(id)
	root, _ := sh.Resolve("/ipns/" + id.ID)
	ls, _ := sh.FileList(root)
	t.Log(ls)
	for _, x := range ls.Links {
		t.Log(x)
	}
}
