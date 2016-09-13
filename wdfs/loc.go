package wdfs

import (
	"encoding/xml"
	"github.com/kpmy/mipfs/ipfs_api"
	. "github.com/kpmy/ypk/tc"
	"golang.org/x/net/webdav"
	"log"
	"os"
	"strings"
)

type loc struct {
	ch *chain
}

func (l *loc) Close() error {
	return nil
}

func (l *loc) Read(p []byte) (n int, err error) {
	return 0, webdav.ErrForbidden
}

func (l *loc) Seek(offset int64, whence int) (int64, error) {
	return 0, webdav.ErrForbidden
}

func (l *loc) Readdir(count int) (ret []os.FileInfo, err error) {
	ls, _ := ipfs_api.Shell().FileList(l.ch.Hash)
	for _, lo := range ls.Links {
		switch lo.Type {
		case "File":
			fallthrough
		case "Directory":
			if !strings.HasPrefix(lo.Name, "*") {
				filepath := l.ch.Hash + "/" + lo.Name
				ret = append(ret, newChain(l.ch, filepath).tail())
			}
		default:
			Halt(100)
		}
	}
	return
}

func (l *loc) Stat() (os.FileInfo, error) {
	return l.ch, nil
}

func (l *loc) Write(p []byte) (n int, err error) {
	return 0, webdav.ErrForbidden
}

func (l *loc) DeadProps() (map[xml.Name]webdav.Property, error) {
	log.Println("loc prop get")
	return nil, nil
}

func (l *loc) Patch(patch []webdav.Proppatch) ([]webdav.Propstat, error) {
	log.Println("loc prop patch", patch)
	return nil, nil
}
