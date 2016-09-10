package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/kpmy/mipfs/ipfs_api"
	"github.com/kpmy/mipfs/wdfs"
	"github.com/peterbourgon/diskv"
	"golang.org/x/net/webdav"
)

var KV *diskv.Diskv

func init() {
	log.SetFlags(0)

	KV = diskv.New(diskv.Options{
		BasePath: ".diskv",
		Transform: func(s string) []string {
			return []string{}
		},
	})
}

func main() {
	root := "QmbuSdtGUUfL7DSvvA9DmiGSRqAzkHEjWtsxZDRPBWcawg"
	if r, err := KV.Read("root"); err == nil {
		root = string(r)
	} else {
		KV.Write("root", []byte(root))
	}
	nodeID, _ := ipfs_api.Shell().ID()
	fs := wdfs.NewFS(nodeID, root)
	ls := wdfs.NewLS(fs)
	h := &webdav.Handler{
		Prefix:     "/ipfs",
		FileSystem: fs,
		LockSystem: ls,
		Logger: func(r *http.Request, err error) {
			switch r.Method {
			case "COPY", "MOVE":
				dst := ""
				if u, err := url.Parse(r.Header.Get("Destination")); err == nil {
					dst = u.Path
				}
				o := r.Header.Get("Overwrite")
				log.Println(r.Method, r.URL.Path, dst, o, err)
			default:
				log.Println(r.Method, r.URL.Path, err)
			}
			KV.Write("root", []byte(fs.String()))
		},
	}
	http.Handle("/ipfs/", h)
	http.HandleFunc("/ipfs", func(resp http.ResponseWriter, req *http.Request) {
		if r, err := KV.Read("root"); err == nil {
			resp.Write(r)
		}
	})
	const addr = "0.0.0.0:6001"
	log.Println("webdav server started at", addr)
	http.ListenAndServe(addr, nil)
}
