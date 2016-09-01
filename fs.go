package mipfs

import (
	"os"
	"time"

	ipfs_api "github.com/ipfs/go-ipfs-api"

	ipfs "github.com/kpmy/mipfs/ipfs_api"
	. "github.com/kpmy/ypk/tc"
	"golang.org/x/net/webdav"
)

type dir struct {
	ipfs_api.UnixLsObject
}

func (d *dir) Name() string {
	return d.UnixLsObject.Hash
}

func (d *dir) Size() int64 {
	return 0
}

func (d *dir) Mode() os.FileMode { return os.ModeDir }

func (d *dir) ModTime() time.Time {
	return time.Now()
}

func (d *dir) IsDir() bool      { return true }
func (d *dir) Sys() interface{} { return nil }

func (d *dir) Close() error {
	return nil
}

func (d *dir) Read(p []byte) (n int, err error) {
	panic(100)
}

func (d *dir) Seek(offset int64, whence int) (int64, error) {
	panic(100)
}

func (d *dir) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (d *dir) Stat() (os.FileInfo, error) {
	ls, _ := ipfs.Shell().FileList(d.Hash)
	return &dir{*ls}, nil
}

func (d *dir) Write(p []byte) (n int, err error) {
	panic(100)
}

type filesystem struct {
	node string
	root string
}

func (f *filesystem) Mkdir(name string, perm os.FileMode) error {
	Assert(name != "", 20)
	panic(100)
}

func (f *filesystem) OpenFile(name string, flag int, perm os.FileMode) (webdav.File, error) {
	ls, _ := ipfs.Shell().FileList(f.root)
	return &dir{*ls}, nil
}

func (f *filesystem) RemoveAll(name string) error {
	panic(100)
}

func (f *filesystem) Rename(oldName, newName string) error {
	panic(100)
}

func (f *filesystem) Stat(name string) (os.FileInfo, error) {
	ls, _ := ipfs.Shell().FileList(f.root)
	return &dir{*ls}, nil
}

var nodeID *ipfs_api.IdOutput;

func init()  {
	nodeID, _ = ipfs.Shell().ID()
}

func NewFS() webdav.FileSystem {
	//root, _ := ipfs.Shell().Resolve(nodeID.ID)
	root := "QmbuSdtGUUfL7DSvvA9DmiGSRqAzkHEjWtsxZDRPBWcawg"
	return &filesystem{node: nodeID.ID, root: root}
}

type locksystem struct {
}

func (l *locksystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	panic(100)
}

func (l *locksystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	panic(100)
}

func (l *locksystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	panic(100)
}

func (l *locksystem) Unlock(now time.Time, token string) error {
	panic(100)
}

func NewLS() webdav.LockSystem {
	return &locksystem{}
}
