package wdfs

import (
	"github.com/ipfs/go-ipfs-api"
	"github.com/kpmy/mipfs/ipfs_api"
	. "github.com/kpmy/ypk/tc"
	"os"
	"strings"
	"time"
)

type chain struct {
	up, down, link *chain
	shell.UnixLsObject
	name string
}

func newChain(root *chain, filepath string) (ret *chain) {
	ns := strings.Split(strings.Trim(filepath, "/"), "/")
	Assert(ns[0] == root.Hash, 20)
	root.name = root.Hash
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

func (root *chain) update(hash string) {
	Assert(hash != "", 20)
	root.Hash = hash
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
	Halt(100)
	return nil
}
