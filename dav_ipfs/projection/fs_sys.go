package projection //import "github.com/kpmy/mipfs/dav_ipfs/projection"

import (
	"fmt"
	"github.com/kpmy/mipfs/dav_ipfs"
	"github.com/kpmy/mipfs/ipfs_api"
	"github.com/streamrail/concurrent-map"
	"golang.org/x/net/webdav"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const Active = true

const Passive = false

const ProjectionRoot = ".fs"

type Extension interface {
	ConnectTo(root string)
	Open(chain []string) (webdav.File, error)
	os.FileInfo
	webdav.File
}

type item struct {
	name string
}

func (i *item) Size() int64 {
	return 0
}
func (i *item) Mode() os.FileMode {
	return 0
}
func (i *item) ModTime() time.Time {
	return time.Now()
}
func (i *item) IsDir() bool {
	return false
}
func (i *item) Sys() interface{} {
	return nil
}
func (i *item) Close() error {
	return nil
}

func (i *item) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (i *item) Seek(offset int64, whence int) (int64, error) {
	log.Println("seek")
	return 0, nil
}

func (i *item) Write(p []byte) (n int, err error) {
	return 0, webdav.ErrForbidden
}
func (i *item) Readdir(count int) (ret []os.FileInfo, err error) {
	return nil, webdav.ErrForbidden
}

func (i *item) Stat() (os.FileInfo, error) {
	return i, nil
}

func (i *item) Name() string {
	return i.name
}

type locator struct {
}

func (l *locator) Size() int64 {
	return 0
}
func (l *locator) Mode() os.FileMode {
	return os.ModeDir
}
func (l *locator) ModTime() time.Time {
	return time.Now()
}
func (l *locator) IsDir() bool {
	return true
}
func (l *locator) Sys() interface{} {
	return nil
}
func (l *locator) Close() error {
	return nil
}
func (l *locator) Read(p []byte) (n int, err error) {
	return 0, webdav.ErrForbidden
}
func (l *locator) Seek(offset int64, whence int) (int64, error) {
	return 0, webdav.ErrForbidden
}
func (l *locator) Write(p []byte) (n int, err error) {
	return 0, webdav.ErrForbidden
}

type ext struct {
	locator
	root string
	name string
}

func (e *ext) Name() string {
	return e.name
}

func (e *ext) Readdir(count int) (ret []os.FileInfo, err error) {
	log.Println(e.root)
	var ch <-chan string
	if ch, err = ipfs_api.Shell().Refs(e.root, true); err == nil {
		for s := range ch {
			ret = append(ret, &item{name: s})
		}
	}
	return
}

func (e *ext) Stat() (os.FileInfo, error) {
	return e, nil
}

func (e *ext) Open(chain []string) (ret webdav.File, err error) {
	if len(chain) == 1 {
		i := &item{name: chain[0]}
		ret = i
	} else {
		err = os.ErrPermission
	}
	return
}

func (e *ext) ConnectTo(root string) {
	e.root = root
}

type cat struct {
	locator
	pl map[string]Extension
}

func (c *cat) Name() string {
	return ProjectionRoot
}

func (c *cat) Readdir(count int) (ret []os.FileInfo, err error) {
	for _, v := range c.pl {
		ret = append(ret, v)
	}
	return
}

func (c *cat) Stat() (os.FileInfo, error) {
	return c, nil
}

type projection struct {
	inner webdav.FileSystem
	root  string
	cache cmap.ConcurrentMap
	all   *cat

	set func(string)
	get func() string
}

func isProjection(split []string) bool {
	return strings.ToLower(split[0]) == ProjectionRoot
}

func (p *projection) Mkdir(name string, perm os.FileMode) (err error) {
	ls := splitPath(name)
	switch {
	case len(ls) > 1 && isProjection(ls):
		err = os.ErrPermission
	case len(ls) == 1 && isProjection(ls):
		err = p.inner.Mkdir(strings.ToLower(name), perm)
	default:
		err = p.inner.Mkdir(name, perm)
	}
	return
}

func (p *projection) OpenFile(name string, flag int, perm os.FileMode) (ret webdav.File, err error) {
	log.Println("open", name)
	ls := splitPath(name)
	switch {
	case isProjection(ls) && len(ls) == 1:
		ret = p.all
	case isProjection(ls) && len(ls) > 1:
		if e, ok := p.all.pl[ls[1]]; ok {
			e.ConnectTo(p.root)
			if len(ls) > 2 {
				ret, err = e.Open(ls[2:])
			} else {
				ret = e
			}
		} else {
			err = os.ErrNotExist
		}
	default:
		ret, err = p.inner.OpenFile(name, flag, perm)
	}
	return
}

func (p *projection) RemoveAll(name string) (err error) {
	ls := splitPath(name)
	switch {
	case isProjection(ls):
		err = os.ErrPermission
	default:
		err = p.inner.RemoveAll(name)
	}
	return
}

func (p *projection) Rename(oldName, newName string) (err error) {
	ls := splitPath(oldName)
	switch {
	case isProjection(ls):
		err = os.ErrPermission
	default:
		err = p.inner.Rename(oldName, newName)
	}
	return
}

func (p *projection) Stat(name string) (fi os.FileInfo, err error) {
	log.Println("stat", name)
	ls := splitPath(name)
	switch {
	case isProjection(ls):
		fi = &cat{}
	default:
		fi, err = p.inner.Stat(name)
	}
	return
}

func (p *projection) String() string {
	return fmt.Sprint(p.inner)
}

func NewPS(get func() string, set func(string), active bool) (fs webdav.FileSystem, ls webdav.LockSystem) {
	var pr *projection
	if active {
		pr = &projection{get: get}
		pr.set = func(hash string) {
			pr.root = hash
			set(hash)
		}
		xs := dav_ipfs.NewFS(pr.get, pr.set)
		ls = dav_ipfs.NewLS(xs)
		pr.inner = xs

		pr.cache = cmap.New()

		pr.all = &cat{}
		pr.all.pl = map[string]Extension{"all": &ext{name: "all"}}

		fs = pr
	} else {
		xs := dav_ipfs.NewFS(get, set)
		ls = dav_ipfs.NewLS(xs)
		fs = xs
	}
	return
}
