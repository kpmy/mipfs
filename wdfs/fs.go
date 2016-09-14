package wdfs

import (
	"github.com/ipfs/go-ipfs-api"
	"github.com/kpmy/mipfs/ipfs_api"
	"github.com/kpmy/ypk/dom"
	"github.com/kpmy/ypk/fn"
	. "github.com/kpmy/ypk/tc"
	"golang.org/x/net/webdav"
	"log"
	"os"
	"strings"
)

type filesystem struct {
	webdav.FileSystem
	nodeId *shell.IdOutput
	root   *chain
}

func (f *filesystem) Mkdir(name string, perm os.FileMode) (err error) {
	chain := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(name, "/"))
	if tail := chain.tail(); !tail.exists() {
		onlyOne := true
		for tail != nil {
			if !tail.exists() {
				if onlyOne {
					tail.Hash, _ = ipfs_api.Shell().NewObject("unixfs-dir")
					prop := newProps()
					propHash, _ := ipfs_api.Shell().Add(dom.EncodeWithHeader(prop))
					if tail.Hash, err = ipfs_api.Shell().PatchLink(tail.Hash, "*", propHash, false); err != nil {
						log.Fatal(err)
						return
					}
					onlyOne = false
				} else {
					err = os.ErrNotExist
					return
				}
			}
			if tail.down != nil {
				tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, tail.down.name, tail.down.Hash, false)
			}
			tail = tail.up
		}
		chain.link.update(chain.Hash)
	} else {
		err = os.ErrExist
	}
	return
}

func (f *filesystem) OpenFile(name string, flag int, perm os.FileMode) (ret webdav.File, err error) {
	log.Println("open", name, flag, perm)
	path := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(name, "/"))
	switch tail := path.tail(); {
	case tail.exists() && tail.IsDir():
		ret = &loc{ch: tail}
	case tail.exists() && !tail.IsDir():
		ret = &file{ch: tail}
	case !tail.exists() && flag&os.O_CREATE != 0:
		var edir *chain
		for edir = tail.up; edir != nil && edir.exists() && edir.IsDir(); edir = edir.up {
		}
		if fn.IsNil(edir) {
			ret = &file{ch: tail}
		} else {
			err = os.ErrNotExist
		}
	default:
		log.Println("open error", name, flag, perm)
		err = os.ErrNotExist
	}
	return
}

func (f *filesystem) RemoveAll(name string) (err error) {
	chain := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(name, "/"))
	if tail := chain.tail(); tail.exists() {
		tail = tail.up
		tail.Hash, _ = ipfs_api.Shell().Patch(tail.Hash, "rm-link", tail.down.name)
		if !tail.down.IsDir() {
			//удалим пропы
			if th, err := ipfs_api.Shell().Patch(tail.Hash, "rm-link", "*"+tail.down.name); err == nil {
				tail.Hash = th
			}
		}
		tail = tail.up
		for tail != nil {
			tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, tail.down.name, tail.down.Hash, false)
			tail = tail.up
		}
		chain.link.update(chain.Hash)
	}
	return
}

func (f *filesystem) Rename(oldName, newName string) (err error) {
	log.Println("rename", oldName, newName)
	on := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(oldName, "/"))
	var op *chain
	if !on.tail().IsDir() {
		propPath := ""
		for x := on; x != nil; x = x.down {
			if x.down == nil {
				propPath = propPath + "/" + "*" + x.name
			} else {
				propPath = propPath + "/" + x.name
			}
		}
		op = newChain(f.root.mirror(), propPath)
	}
	nn := newChain(f.root.mirror(), f.root.Hash+"/"+strings.Trim(newName, "/"))
	if ot := on.tail(); ot.exists() {
		if nt := nn.tail(); !nt.exists() {
			Assert(ot.depth() == nt.depth(), 40)
			tail := ot.up
			tail.Hash, _ = ipfs_api.Shell().Patch(tail.Hash, "rm-link", ot.name)
			if op != nil {
				tail.Hash, _ = ipfs_api.Shell().Patch(tail.Hash, "rm-link", op.tail().name)
			}
			tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, nt.name, ot.Hash, false)
			if op != nil {
				tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, "*"+nt.name, op.tail().Hash, false)
			}
			tail = tail.up
			for tail != nil {
				tail.Hash, _ = ipfs_api.Shell().PatchLink(tail.Hash, tail.down.name, tail.down.Hash, false)
				tail = tail.up
			}
			on.link.update(on.Hash)
		} else {
			err = os.ErrExist
		}
	} else {
		err = os.ErrNotExist
	}
	return
}

func (f *filesystem) Stat(name string) (fi os.FileInfo, err error) {
	log.Println("stat", name)
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
