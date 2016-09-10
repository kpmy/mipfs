package wdfs

import (
	"fmt"
	"github.com/ipfs/go-ipfs-api"
	"github.com/kpmy/mipfs/ipfs_api"
	. "github.com/kpmy/ypk/tc"
	"github.com/mattetti/filebuffer"
	"golang.org/x/net/webdav"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
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
			filepath := l.ch.Hash + "/" + lo.Name
			ret = append(ret, newChain(l.ch, filepath).tail())
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

type block struct {
	pos  int64
	data *filebuffer.Buffer
}

type file struct {
	ch    *chain
	pos   int64
	links []*shell.LsLink
	buf   *filebuffer.Buffer

	wr chan *block
	wg *sync.WaitGroup
}

func (f *file) Name() string {
	return f.ch.name
}

func (f *file) Size() int64 {
	return f.ch.Size()
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
	if f.wr != nil {
		close(f.wr)
		f.wg.Wait()
	}
	return nil
}

func (f *file) Read(p []byte) (n int, err error) {
	if f.links == nil {
		f.links, _ = ipfs_api.Shell().List(f.ch.Hash)
	}
	if len(f.links) == 0 {
		buf := filebuffer.New(nil)
		rd, _ := ipfs_api.Shell().Cat(f.ch.Hash)
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
	return nil, webdav.ErrForbidden
}

func (f *file) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *file) update(data io.ReadCloser) {
	f.ch.Hash, _ = ipfs_api.Shell().Add(data)
	for tail := f.ch.up; tail != nil; tail = tail.up {
		tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, tail.down.name, tail.down.Hash, false)
	}
	head := f.ch.head()
	head.link.Hash = head.Hash
}

func (f *file) Write(p []byte) (n int, err error) {
	if f.wr == nil {
		f.wr = make(chan *block, 16)
		f.wg = new(sync.WaitGroup)
		f.wg.Add(1)
		go func(f *file) {
			tmp, _ := ioutil.TempFile(os.TempDir(), "mipfs")
			for w := range f.wr {
				tmp.Seek(w.pos, io.SeekStart)
				w.data.Seek(0, io.SeekStart)
				io.Copy(tmp, w.data)
			}
			tmp.Seek(0, io.SeekStart)
			f.update(tmp)
			f.wg.Done()
		}(f)
	}
	b := &block{pos: f.pos}
	b.data = filebuffer.New(nil)
	n, err = b.data.Write(p)
	f.wr <- b
	f.pos = f.pos + int64(n)
	return n, nil
}

type chain struct {
	up, down, link *chain
	shell.UnixLsObject
	name string
}

func newChain(root *chain, filepath string) (ret *chain) {
	ns := strings.Split(strings.Trim(filepath, "/"), "/")
	Assert(ns[0] == root.Hash, 20)
	ret = root
	for i := 1; i < len(ns); i++ {
		down := &chain{}
		down.name = ns[i]
		down.up, root.down = root, down
		root = down
	}
	skip := false
	for root = ret; root != nil; root = root.down {
		if !skip {
			if o, err := ipfs_api.Shell().FileList(root.Hash); err == nil {
				root.UnixLsObject = *o
				if root.down != nil {
					skip = true
					for _, l := range o.Links {
						if l.Name == root.down.name {
							skip = false
							root.down.Hash = l.Hash
						}
					}
				}
			} else {
				Halt(100, root.name, err)
			}
		}
	}
	return
}

func (root *chain) tail() (ret *chain) {
	for ret = root; ret.down != nil; ret = ret.down {
	}
	return
}

func (root *chain) head() (ret *chain) {
	for ret = root; ret.up != nil; ret = ret.up {
	}
	return
}

func (root *chain) exists() bool {
	return root.Hash != ""
}

func (root *chain) depth() (ret int) {
	for tail := root; tail != nil; tail = tail.up {
		ret++
	}
	return
}

func (root *chain) mirror() (ret *chain) {
	Assert(root.up == nil && root.down == nil, 20)
	ret = &chain{}
	ret.Hash = root.Hash
	ret.link = root
	return
}

func (c *chain) Name() string {
	return c.name
}

func (c *chain) Size() int64 {
	return int64(c.UnixLsObject.Size)
}

func (c *chain) Mode() os.FileMode {
	if c.Type == "Directory" {
		return os.ModeDir
	} else if c.Type == "File" {
		return 0
	}
	panic(100)
}

func (c *chain) ModTime() time.Time {
	return time.Now()
}

func (c *chain) IsDir() bool {
	return c.Mode() == os.ModeDir
}
func (c *chain) Sys() interface{} {
	return nil
}

type filesystem struct {
	webdav.FileSystem
	nodeId *shell.IdOutput
	root   *chain
}

func (f *filesystem) Mkdir(name string, perm os.FileMode) (err error) {
	chain := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(name, "/"))
	if tail := chain.tail(); !tail.exists() {
		for tail != nil {
			if !tail.exists() {
				tail.Hash, _ = ipfs_api.Shell().NewObject("unixfs-dir")
			}
			if tail.down != nil {
				tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, tail.down.name, tail.down.Hash, false)
			}
			tail = tail.up
		}
		chain.link.Hash = chain.Hash
	} else {
		err = os.ErrExist
	}
	return
}

func (f *filesystem) OpenFile(name string, flag int, perm os.FileMode) (ret webdav.File, err error) {
	chain := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(name, "/"))
	switch tail := chain.tail(); {
	case tail.exists() && tail.IsDir():
		ret = &loc{ch: tail}
	case tail.exists() && !tail.IsDir():
		ret = &file{ch: tail}
	case !tail.exists() && flag&os.O_CREATE != 0:
		ret = &file{ch: tail}
	default:
		log.Println(name, flag, perm)
		err = os.ErrNotExist
	}
	return
}

func (f *filesystem) RemoveAll(name string) (err error) {
	chain := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(name, "/"))
	if tail := chain.tail(); tail.exists() {
		tail = tail.up
		tail.Hash, _ = ipfs_api.Shell().Patch(tail.Hash, "rm-link", tail.down.name)
		tail = tail.up
		for tail != nil {
			tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, tail.down.name, tail.down.Hash, false)
			tail = tail.up
		}
		chain.link.Hash = chain.Hash
	}
	return
}

func (f *filesystem) Rename(oldName, newName string) (err error) {
	on := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(oldName, "/"))
	nn := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(newName, "/"))
	if ot := on.tail(); ot.exists() {
		if nt := nn.tail(); !nt.exists() {
			Assert(ot.depth() == nt.depth(), 40)
			tail := ot.up
			tail.Hash, _ = ipfs_api.Shell().Patch(tail.Hash, "rm-link", ot.name)
			tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, nt.name, ot.Hash, false)
			tail = tail.up
			for tail != nil {
				tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, tail.down.name, tail.down.Hash, false)
				tail = tail.up
			}
			on.link.Hash = on.Hash
		} else {
			err = os.ErrExist
		}
	} else {
		err = os.ErrNotExist
	}
	return
}

func (f *filesystem) Stat(name string) (fi os.FileInfo, err error) {
	chain := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(name, "/"))
	tail := chain.tail()
	if fi = tail; !tail.exists() {
		err = os.ErrNotExist
	}
	return
}

func (f *filesystem) String() string {
	return f.root.Hash
}

func NewFS(id *shell.IdOutput, root string) *filesystem {
	ch := &chain{}
	ch.Hash = root
	ch.name = root
	ch.Type = "Directory"
	return &filesystem{nodeId: id, root: ch}
}

type locksystem struct {
	webdav.LockSystem
	sync.RWMutex
	locks  map[string]string
	tokens map[string]webdav.LockDetails
}

func (l *locksystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	l.RLock()
	if _, ok := l.locks[name0]; ok {
		release = func() {
			log.Println(name0, "release")
		}
	} else {
		err = webdav.ErrConfirmationFailed
	}
	l.RUnlock()
	return
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

func NewLS(fs webdav.FileSystem) *locksystem {
	ret := &locksystem{}
	ret.locks = make(map[string]string)
	ret.tokens = make(map[string]webdav.LockDetails)
	return ret
}
