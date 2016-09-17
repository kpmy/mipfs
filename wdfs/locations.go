package wdfs

import (
	"encoding/xml"
	"github.com/kpmy/mipfs/ipfs_api"
	"github.com/kpmy/ypk/dom"
	"github.com/kpmy/ypk/fn"
	. "github.com/kpmy/ypk/tc"
	"golang.org/x/net/webdav"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type loc struct {
	ch    *chain
	props dom.Element
}

func (l *loc) Name() string {
	return l.ch.name
}

func (l *loc) Size() int64 {
	return int64(l.ch.UnixLsObject.Size)
}

func (l *loc) Mode() os.FileMode {
	if l.ch.Type == "Directory" {
		return os.ModeDir
	} else if l.ch.Type == "File" {
		return 0
	}
	panic(100)
}

func (l *loc) ModTime() (ret time.Time) {
	ret = time.Now()
	if !fn.IsNil(l.props) {
		if ts := l.props.Attr("modified"); ts != "" {
			if sec, err := strconv.ParseInt(ts, 10, 64); err == nil {
				ret = time.Unix(sec, 0)
			}
		}
	}
	return
}

func (l *loc) IsDir() bool {
	return true
}

func (l *loc) Sys() interface{} {
	Halt(100)
	return nil
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
				tail := newChain(l.ch, filepath).tail()
				var fi os.FileInfo
				if tail.IsDir() {
					_l := &loc{ch: tail}
					_l.readPropsModel()
					fi = _l
				} else {
					_f := &file{ch: tail}
					_f.readPropsModel()
					fi = _f
				}
				ret = append(ret, fi)
			}
		default:
			Halt(100)
		}
	}
	return
}

func (l *loc) Stat() (os.FileInfo, error) {
	return l, nil
}

func (l *loc) Write(p []byte) (n int, err error) {
	return 0, webdav.ErrForbidden
}

func (l *loc) readPropsModel() {
	ls, _ := ipfs_api.Shell().FileList(l.ch.Hash)
	pm := propLinksMap(ls)
	if p, ok := pm["*"]; ok {
		rd, _ := ipfs_api.Shell().Cat(p.Hash)
		if el, err := dom.Decode(rd); err == nil {
			l.props = el.Model()
		} else {
			Halt(99)
		}
	} else {
		l.props = newPropsModel()
	}
}

func (l *loc) readPropsObject() (props map[xml.Name]dom.Element, err error) {
	l.readPropsModel()
	props = make(map[xml.Name]dom.Element)
	props = readProps(l.props)
	return
}

func (l *loc) writePropsObject(props map[xml.Name]dom.Element) {
	el := writeProps(props)
	propHash, _ := ipfs_api.Shell().Add(dom.EncodeWithHeader(el))
	for tail := l.ch; tail != nil; tail = tail.up {
		if tail.Hash == l.ch.Hash {
			tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, "*", propHash, false)
		} else {
			tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, tail.down.name, tail.down.Hash, false)
		}
	}
	head := l.ch.head()
	head.link.update(head.Hash)
}

func (l *loc) DeadProps() (ret map[xml.Name]webdav.Property, err error) {
	log.Println("loc props get")
	pm, _ := l.readPropsObject()
	ret = props2webdav(pm)
	log.Println(ret)
	return
}

func (l *loc) Patch(patch []webdav.Proppatch) (ret []webdav.Propstat, err error) {
	log.Println("loc prop patch", patch)
	pe, _ := l.readPropsObject()
	ret = propsPatch(pe, patch)
	log.Println("loc file props", pe)
	l.writePropsObject(pe)
	return

}
