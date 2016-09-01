package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/kpmy/mipfs"
	"golang.org/x/net/webdav"
)

func init()  {
	log.SetFlags(0);
}

func main() {
	fs := mipfs.NewFS()
	ls := mipfs.NewLS()
	h := &webdav.Handler{
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
				log.Printf("%-20s%-10s%-30s%-30so=%-2s%v", r.Method, r.URL.Path, dst, o, err)
			default:
				log.Printf("%-20s%-10s%-30s%v", r.Method, r.URL.Path, err)
			}
		},
	}
	http.Handle("/", h)
	const addr = "0.0.0.0:6001"
	log.Println("webdav server started at", addr)
	http.ListenAndServe(addr, nil)
}
