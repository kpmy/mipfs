package main

import (
	"bufio"
	"bytes"
	"github.com/abbot/go-http-auth"
	"github.com/kpmy/mipfs/dav_ipfs"
	"github.com/kpmy/mipfs/dav_ipfs/projection"
	"github.com/kpmy/mipfs/ipfs_api"
	. "github.com/kpmy/ypk/tc"
	"github.com/streamrail/concurrent-map"
	"github.com/tv42/zbase32"
	"golang.org/x/net/webdav"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var um cmap.ConcurrentMap = cmap.New()
var rm cmap.ConcurrentMap = cmap.New()

type FileLockSys struct {
	fs  webdav.FileSystem
	ls  webdav.LockSystem
	dav *webdav.Handler
}

const HistoryLimit = 256

type pin struct {
	hash string
	pin  bool
}

var importantHash map[string]string = map[string]string{dav_ipfs.EmptyDirHash: "empty unixfs dir", dav_ipfs.EmptyFileHash: "empty unixfs file"}

func writeRoot(ch chan string, user string) {
	pinCh := make(chan pin, 1024)
	go func() {
		for p := range pinCh {
			if p.pin {
				ipfs_api.Shell().Pin(p.hash)
				//log.Println("pin", p.hash)
			} else if _, ok := importantHash[p.hash]; !ok {
				ipfs_api.Shell().Unpin(p.hash)
				//log.Println("unpin", p.hash)
			}
		}
	}()
	for {
		for s := range ch {
			if s != "" {
				if old, err := KV.Read(user + ".root"); err == nil && s != string(old) {
					history := new(bytes.Buffer)
					history.Write(old)
					history.Write([]byte("\n"))
					if hs, err := KV.Read(user + ".root.history"); err == nil {
						oldHistory := bytes.NewBuffer(hs)
						io.CopyN(history, oldHistory, int64(history.Len()*HistoryLimit)) //лимит истории
						rd := bufio.NewReader(oldHistory)
						for {
							if us, err := rd.ReadString('\n'); err == nil && us != "" {
								us = strings.TrimSpace(us)
								pinCh <- pin{hash: us}
							} else {
								break
							}
						}
					}
					KV.Write(user+".root.history", history.Bytes())
				}
				KV.Write(user+".root", []byte(s))
				pinCh <- pin{hash: s, pin: true}
			} else {
				Halt(100, "empty root")
			}
		}
	}
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
			user := zbase32.EncodeToString([]byte(req.Username))
			fl := new(FileLockSys)

			rootWr := make(chan string, 1024)
			go writeRoot(rootWr, user)

			defaultRoot := dav_ipfs.EmptyDirHash
			if r, err := KV.Read(user + ".root"); err == nil && len(r) > 0 {
				defaultRoot = string(r)
				found := make(chan string)
				go func() {
					if _, err := ipfs_api.Shell().BlockGet(defaultRoot); err != nil {
						defaultRoot = dav_ipfs.EmptyDirHash
					}
					found <- defaultRoot
				}()
				select {
				case <-found:
				case <-time.After(10 * time.Second):
					defaultRoot = dav_ipfs.EmptyDirHash
				}
			} else {
				KV.Write(user+".root", []byte(defaultRoot))
			}
			rm.Set(user+".root", defaultRoot)
			fl.fs, fl.ls = projection.NewPS(func() string {
				r, _ := rm.Get(user + ".root")
				return r.(string)
			}, func(hash string) {
				rm.Set(user+".root", hash)
				rootWr <- hash
			}, projection.Active)
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
