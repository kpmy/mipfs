package ipfs_api

import (
	"github.com/ipfs/go-ipfs-api"
	. "github.com/kpmy/ypk/tc"
	"github.com/mattetti/filebuffer"
	"io"
)

func (sh *MyShell) FileList(hash string) (ret *shell.UnixLsObject, err error) {
	if x, ok := sh.cache.Get(hash); !ok {
		if ret, err = sh.Shell.FileList(hash); err == nil {
			sh.cache.Set(hash, ret)
		}
	} else {
		ret = x.(*shell.UnixLsObject)
	}
	return
}

const BufferLimit = 2048

func (sh *MyShell) CacheCat(hash string) (ret io.ReadCloser, err error) {
	if x, ok := sh.cache.Get(hash); !ok {
		var old io.ReadCloser
		if old, err = sh.Shell.Cat(hash); err == nil {
			buf := filebuffer.New(nil)
			io.CopyN(buf, old, BufferLimit+1)
			Assert(buf.Buff.Len() <= BufferLimit, 40, "buffer too large")
			buf.Seek(0, io.SeekStart)
			sh.cache.Set(hash, buf)
			ret = buf
		}
	} else {
		buf := x.(*filebuffer.Buffer)
		buf.Seek(0, io.SeekStart)
		ret = buf
	}
	return
}
