package main

import (
	"github.com/abbot/go-http-auth"
	"github.com/kpmy/mipfs/ipfs_api"
	"github.com/kpmy/mipfs/wdfs"
	. "github.com/kpmy/ypk/tc"
	"github.com/peterbourgon/diskv"
	"github.com/tv42/zbase32"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var KV *diskv.Diskv

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	KV = diskv.New(diskv.Options{
		BasePath: ".diskv",
		Transform: func(s string) []string {
			return []string{}
		},
	})
}

func main() {
	dir, _ := os.Getwd()
	log.Println("started at", dir)

	if r, err := KV.Read("ipfs"); err == nil {
		ipfs_api.Addr = string(r)
	}

	log.Println("ipfs api at", ipfs_api.Addr)
	if _, err := ipfs_api.Shell().ID(); err == nil {
		Assert(ipfs_api.Shell().Pin(wdfs.EmptyDirHash) == nil && ipfs_api.Shell().Pin(wdfs.EmptyFileHash) == nil, 40)
	} else {
		log.Fatal(err)
	}
	dav := handler()
	http.Handle("/ipfs/", dav)
	http.Handle("/ipfs", dav)

	http.HandleFunc("/hash", auth.NewBasicAuthenticator("ipfs", func(user, realm string) (ret string) {
		un := zbase32.EncodeToString([]byte(user))
		if hash, err := KV.Read(un); err == nil {
			ret = string(hash)
		}
		return
	}).Wrap(func(resp http.ResponseWriter, req *auth.AuthenticatedRequest) {
		un := zbase32.EncodeToString([]byte(req.Username))
		if rh, err := KV.Read(un + ".root"); err == nil {
			rootHash := string(rh)
			tpl := "<html><body><a href='https://ipfs.io/ipfs/" + rootHash + "' target='_blank'>/ipfs/" + rootHash + "</a></body></html>"
			resp.Write([]byte(tpl))
		}
	}))

	http.Handle("/user", regHandler())

	const addr = "0.0.0.0:6001"
	log.Println("webdav server started at", addr)
	log.Println(http.ListenAndServe(addr, nil))
}
