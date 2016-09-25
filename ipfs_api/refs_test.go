package ipfs_api

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/ipfs/go-ipfs-api"
	"github.com/kpmy/ypk/fn"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

func TestRefsLocal(t *testing.T) {
	//sh := shell.NewShell("127.0.0.1:5001")
	var url = "127.0.0.1:5001"
	if a, err := ma.NewMultiaddr(url); err == nil {
		_, host, err := manet.DialArgs(a)
		if err == nil {
			url = host
		}
	}
	req := shell.NewRequest(url, "refs/local")
	if resp, err := req.Send(http.DefaultClient); err == nil {
		if !fn.IsNil(resp.Error) {
			t.Error(resp.Error)
		} else {
			buf := &bytes.Buffer{}
			io.Copy(buf, resp.Output)
			t.Log(buf.String())
		}
	} else {
		t.Error(err)
	}
	sh := Shell()
	if ch, err := sh.LocalRefs(); err == nil {
		for x := range ch {
			t.Log(x)
		}
	} else {
		t.Error(err)
	}
}
