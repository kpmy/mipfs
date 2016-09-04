package mipfs

import (
	"io"
	"os"
	"time"

	go_ipfs_api "github.com/ipfs/go-ipfs-api"

	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kpmy/mipfs/ipfs_api"
	. "github.com/kpmy/ypk/tc"
	"github.com/mattetti/filebuffer"
	"golang.org/x/net/webdav"
)

type file struct {
	go_ipfs_api.UnixLsLink
	pos   int64
	buf   *filebuffer.Buffer
	links []*go_ipfs_api.LsLink
}

func (f *file) Name() string {
	return f.UnixLsLink.Name
}

func (f *file) Size() int64 {
	return int64(f.UnixLsLink.Size)
}

func (f *file) Mode() os.FileMode {
	return 0
}

func (f *file) ModTime() time.Time {
	return time.Now()
}

func (f *file) IsDir() bool {
	return false
}
func (f *file) Sys() interface{} {
	return nil
}

func (f *file) Close() error {
	return nil
}

func (f *file) Read(p []byte) (n int, err error) {
	if f.links == nil {
		f.links, _ = ipfs_api.Shell().List(f.Hash)
	}
	if len(f.links) == 0 {
		buf := filebuffer.New(nil)
		rd, _ := ipfs_api.Shell().Cat(f.Hash)
		io.Copy(buf, rd)
		buf.Seek(f.pos, io.SeekStart)
		return buf.Read(p)
	} else {
		var end int64 = 0
		for _, l := range f.links {
			begin := end
			end = begin + int64(l.Size)
			if begin <= f.pos && f.pos < end {
				if f.buf == nil {
					rd, _ := ipfs_api.Shell().Cat(l.Hash)
					f.buf = filebuffer.New(nil)
					io.Copy(f.buf, rd)
					l.Size = uint64(f.buf.Buff.Len())
				}
				f.buf.Seek(f.pos-begin, io.SeekStart)
				n, err = f.buf.Read(p)
				f.pos = f.pos + int64(n)
				if f.buf.Index == int64(l.Size) {
					f.buf = nil
				}
				return
			}
		}
		panic(100)
	}
}

func (f *file) Seek(offset int64, whence int) (seek int64, err error) {
	switch whence {
	case io.SeekStart:
		f.pos = offset
	case io.SeekCurrent:
		f.pos = f.pos + offset
	case io.SeekEnd:
		f.pos = f.Size() + offset
	default:
		Halt(100)
	}
	Assert(f.pos >= 0, 60)
	seek = f.pos
	return
}

func (f *file) Readdir(count int) (ret []os.FileInfo, err error) {
	panic(100)
}

func (f *file) Stat() (os.FileInfo, error) {
	return &link{f.UnixLsLink}, nil
}

func (f *file) Write(p []byte) (n int, err error) {
	panic(100)
}

type link struct {
	go_ipfs_api.UnixLsLink
}

func (l *link) Name() string {
	return l.UnixLsLink.Name
}

func (l *link) Size() int64 {
	return int64(l.UnixLsLink.Size)
}

func (l *link) Mode() os.FileMode {
	if l.Type == "Directory" {
		return os.ModeDir
	} else if l.Type == "File" {
		return 0
	}
	panic(100)
}

func (l *link) ModTime() time.Time {
	return time.Now()
}

func (l *link) IsDir() bool {
	return l.Type == "Directory"
}
func (l *link) Sys() interface{} {
	return nil
}

type loc struct {
	go_ipfs_api.UnixLsObject
}

func (l *loc) Name() string {
	return l.Hash
}

func (l *loc) Size() int64 {
	return int64(l.UnixLsObject.Size)
}

func (l *loc) Mode() os.FileMode {
	return os.ModeDir
}

func (l *loc) ModTime() time.Time {
	return time.Now()
}

func (l *loc) IsDir() bool {
	return true
}
func (l *loc) Sys() interface{} {
	return nil
}

func (l *loc) Close() error {
	return nil
}

func (l *loc) Read(p []byte) (n int, err error) {
	panic(100)
}

func (l *loc) Seek(offset int64, whence int) (int64, error) {
	panic(100)
}

func (l *loc) Readdir(count int) (ret []os.FileInfo, err error) {
	ls, _ := ipfs_api.Shell().FileList(l.Hash)
	for _, l := range ls.Links {
		switch l.Type {
		case "File":
			fallthrough
		case "Directory":
			ret = append(ret, &link{*l})
		default:
			Halt(100)
		}
	}
	return
}

func (l *loc) Stat() (os.FileInfo, error) {
	ls, _ := ipfs_api.Shell().FileList(l.Hash)
	return &loc{*ls}, nil
}

func (l *loc) Write(p []byte) (n int, err error) {
	panic(100)
}

type filesystem struct {
	node string
	root string
}

func (f *filesystem) Mkdir(name string, perm os.FileMode) (err error) {
	Assert(name != "", 20)
	ls := split(f.root, name)
	ns := strings.Split(f.root+name, "/")
	downHash := ""
	downPath := ""
	for i := len(ns) - 1; i >= 0; i-- {
		newHash := ""
		if i >= len(ls) {
			newHash, _ = ipfs_api.Shell().NewObject("unixfs-dir")
		} else {
			newHash = ls[i].Hash
		}
		if downHash != "" {
			newHash, _ = ipfs_api.Shell().PatchLink(newHash, downPath, downHash, false)
		}
		downHash = newHash
		downPath = ns[i]
		if i == 0 {
			ipfs_api.Shell().Unpin(f.root)
			f.root = newHash
			ipfs_api.Shell().Pin(f.root)
			memo.Write("root", []byte(f.root))
		}
	}
	return
}

func (f *filesystem) OpenFile(name string, flag int, perm os.FileMode) (webdav.File, error) {
	if li, fi := trav(f.root, name); fi != nil {
		return &file{UnixLsLink: fi.UnixLsLink}, nil
	} else if li != nil {
		return li, nil
	} else {
		panic(0)
	}
}

func (f *filesystem) RemoveAll(name string) (err error) {
	var ls []*loc
	var ns []string
	var newHash string
	if li, fi := trav(f.root, name); fi != nil {
		ls = split(f.root, filepath.Dir(name))
		ns = strings.Split(f.root+filepath.Dir(name), "/")
		_, fn := filepath.Split(name)
		newHash, _ = ipfs_api.Shell().Patch(ls[len(ls)-1].Hash, "rm-link", fn)
		if j := len(ls) - 2; j > 0 {
			newHash, _ = ipfs_api.Shell().Patch(ls[len(ls)-2].Hash, "rm-link", ns[len(ns)-1])
		}
	} else if li != nil {
		ls = split(f.root, name)
		ns = strings.Split(f.root+name, "/")
		newHash, _ = ipfs_api.Shell().Patch(ls[len(ls)-2].Hash, "rm-link", ns[len(ns)-1])
		Assert(len(ls) > 1 && len(ns) > 1 && len(ls) == len(ns), 20)
	} else {
		panic(0)
	}
	if j := len(ls) - 2; j > 0 {
		for i := j - 1; i >= 0; i-- {
			newHash, _ = ipfs_api.Shell().PatchLink(ls[i].Hash, ns[i+1], newHash, false)
		}
	}
	ipfs_api.Shell().Unpin(f.root)
	f.root = newHash
	ipfs_api.Shell().Pin(f.root)
	memo.Write("root", []byte(f.root))
	return
}

func (f *filesystem) Rename(oldName, newName string) error {
	panic(100)
}

func (f *filesystem) Stat(name string) (os.FileInfo, error) {
	if li, fi := trav(f.root, name); fi != nil {
		return fi, nil
	} else if li != nil {
		return li, nil
	} else {
		panic(0)
	}
}

var nodeID *go_ipfs_api.IdOutput

func init() {
	nodeID, _ = ipfs_api.Shell().ID()
}

func NewFS() webdav.FileSystem {
	//root, _ := ipfs.Shell().Resolve(nodeID.ID)
	root := "QmbuSdtGUUfL7DSvvA9DmiGSRqAzkHEjWtsxZDRPBWcawg"
	if r, err := memo.Read("root"); err == nil {
		root = string(r)
	} else {
		memo.Write("root", []byte(root))
	}
	return &filesystem{node: nodeID.ID, root: root}
}

type locksystem struct {
	sync.RWMutex
	locks  map[string]string
	tokens map[string]webdav.LockDetails
}

func (l *locksystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	panic(100)
}

func (l *locksystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	l.RLock()
	if _, ok := l.locks[details.Root]; !ok {
		l.RUnlock()
		l.Lock()
		token = fmt.Sprint(now.UnixNano())
		l.locks[details.Root] = token
		l.tokens[token] = details
		l.RWMutex.Unlock()
	} else {
		l.RUnlock()
		err = webdav.ErrLocked
	}
	return
}

func (l *locksystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	panic(100)
}

func (l *locksystem) Unlock(now time.Time, token string) (err error) {
	l.Lock()
	details := l.tokens[token]
	delete(l.tokens, token)
	delete(l.locks, details.Root)
	l.RWMutex.Unlock()
	return
}

func NewLS() webdav.LockSystem {
	ret := &locksystem{}
	ret.locks = make(map[string]string)
	ret.tokens = make(map[string]webdav.LockDetails)
	return ret
}
