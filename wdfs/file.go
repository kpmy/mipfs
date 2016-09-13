package wdfs

import (
	"encoding/xml"
	"github.com/ipfs/go-ipfs-api"
	"github.com/kpmy/mipfs/ipfs_api"
	"github.com/kpmy/ypk/fn"
	"github.com/mattetti/filebuffer"
	"io/ioutil"
	"os"
	"sync"
	"time"

	. "github.com/kpmy/ypk/tc"
	"golang.org/x/net/webdav"
	"io"
	"log"
)

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
	Halt(100)
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
		if fn.IsNil(f.buf) {
			f.buf = filebuffer.New(nil)
			rd, _ := ipfs_api.Shell().Cat(f.ch.Hash)
			io.Copy(f.buf, rd)
		}
		f.buf.Seek(f.pos, io.SeekStart)
		n, err = f.buf.Read(p)
		f.pos = f.pos + int64(n)
		return n, err
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

func (f *file) DeadProps() (map[xml.Name]webdav.Property, error) {
	log.Println("file prop get")
	return nil, nil
}

func (f *file) Patch(patch []webdav.Proppatch) ([]webdav.Propstat, error) {
	log.Println("file prop patch", patch)
	return nil, nil
}
