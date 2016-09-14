package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"fmt"
	"os"

	"github.com/kpmy/mipfs/ipfs_api"
	"github.com/kpmy/mipfs/wdfs"
	"github.com/kpmy/ypk/fn"
	. "github.com/kpmy/ypk/tc"
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
	log.Println(os.Getwd())
	root := "QmbuSdtGUUfL7DSvvA9DmiGSRqAzkHEjWtsxZDRPBWcawg"
	if r, err := KV.Read("root"); err == nil && len(r) > 0 {
		root = string(r)
	} else {
		KV.Write("root", []byte(root))
	}

	if r, err := KV.Read("ipfs"); err == nil {
		ipfs_api.Addr = string(r)
	}

	rootCh := make(chan string, 16)
	go func(ch chan string) {
		for {
			i := 0
			for s := range ch {
				if s != "" {
					if old, err := KV.Read("root"); err == nil && s != string(old) {
						KV.Write(fmt.Sprint("root.", time.Now().UnixNano(), ".", i), old)
						i++
					}
					KV.Write("root", []byte(s))
				} else {
					Halt(100, "empty root")
				}
			}
		}
	}(rootCh)

	var fs webdav.FileSystem
	var ls webdav.LockSystem
	if nodeID, err := ipfs_api.Shell().ID(); err == nil {
		fs = wdfs.NewFS(nodeID, root)
		ls = wdfs.NewLS(fs)
	} else {
		log.Fatal(err)
	}
	if !fn.IsNil(fs) {
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
				//log.Println(fs)
				rootCh <- fmt.Sprint(fs)
			},
		}
		http.Handle("/ipfs/", h)
		http.Handle("/ipfs", h)
	}
	http.HandleFunc("/hash", func(resp http.ResponseWriter, req *http.Request) {
		if r, err := KV.Read("root"); err == nil {
			resp.Write(r)
		}
	})
	const addr = "0.0.0.0:6001"
	log.Println("webdav server started at", addr)
	http.ListenAndServe(addr, nil)
}
