package main

import (
	"bytes"
	"github.com/abbot/go-http-auth"
	"github.com/kpmy/mipfs/wdfs"
	. "github.com/kpmy/ypk/tc"
	"github.com/streamrail/concurrent-map"
	"github.com/tv42/zbase32"
	"golang.org/x/net/webdav"
	"io"
	"log"
	"net/http"
	"net/url"
)

var um cmap.ConcurrentMap = cmap.New()

type FileLockSys struct {
	fs  webdav.FileSystem
	ls  webdav.LockSystem
	dav *webdav.Handler
}

func handler() http.Handler {
	return auth.NewBasicAuthenticator("ipfs", func(user, realm string) (ret string) {
		un := zbase32.EncodeToString([]byte(user))
		if hash, err := KV.Read(un); err == nil {
			ret = string(hash)
		}
		return
	}).Wrap(func(resp http.ResponseWriter, req *auth.AuthenticatedRequest) {
		if !um.Has(req.Username) {
			fl := new(FileLockSys)

			rootCh := make(chan string, 256)
			go func(ch chan string, user string) {
				user = zbase32.EncodeToString([]byte(user))
				for {
					i := 0
					for s := range ch {
						if s != "" {
							if old, err := KV.Read(user + ".root"); err == nil && s != string(old) {
								history := new(bytes.Buffer)
								history.Write(old)
								history.Write([]byte("\n"))
								if hs, err := KV.Read(user + ".root.history"); err == nil {
									io.CopyN(history, bytes.NewBuffer(hs), int64(history.Len()*128)) //лимит истории
								}
								KV.Write(user+".root.history", history.Bytes())
								i++
							}
							KV.Write(user+".root", []byte(s))
						} else {
							Halt(100, "empty root")
						}
					}
				}
			}(rootCh, req.Username)

			fs := wdfs.NewFS(func() string {
				user := zbase32.EncodeToString([]byte(req.Username))
				defaultRoot := wdfs.EmptyDirHash
				if r, err := KV.Read(user + ".root"); err == nil && len(r) > 0 {
					defaultRoot = string(r)
				} else {
					KV.Write(user+".root", []byte(defaultRoot))
				}
				return defaultRoot
			}, func(hash string) {
				rootCh <- hash
			})
			fl.fs = fs
			fl.ls = wdfs.NewLS(fs)
			fl.dav = &webdav.Handler{
				Prefix:     "/ipfs",
				FileSystem: fl.fs,
				LockSystem: fl.ls,
				Logger: func(r *http.Request, err error) {
					switch r.Method {
					case "COPY", "MOVE":
						dst := ""
						if u, err := url.Parse(r.Header.Get("Destination")); err == nil {
							dst = u.Path
						}
						o := r.Header.Get("Overwrite")
						log.Println(r.Method, r.URL.Path, dst, o, fl.fs, err)
					default:
						log.Println(r.Method, r.URL.Path, fl.fs, err)
					}
				},
			}

			um.Set(req.Username, fl)
		}
		if fl, ok := um.Get(req.Username); ok {
			fl.(*FileLockSys).dav.ServeHTTP(resp, &req.Request)
		} else {
			Halt(100, "strange user", req.Username)
		}
	})
}
