package mipfs

import (
	"os"
	"time"

	"github.com/ipfs/go-ipfs-api"

	"github.com/kpmy/mipfs/ipfs_api"
	. "github.com/kpmy/ypk/tc"
	"golang.org/x/net/webdav"
)

type file struct {
	go_ipfs_api.UnixLsLink
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
	panic(100)
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	panic(100)
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

func (f *filesystem) Mkdir(name string, perm os.FileMode) error {
	Assert(name != "", 20)
	panic(100)
}

func (f *filesystem) OpenFile(name string, flag int, perm os.FileMode) (webdav.File, error) {
	if li, fi := trav(f.root, name); fi != nil {
		return &file{fi.UnixLsLink}, nil
	} else if li != nil {
		return li, nil
	} else {
		panic(0)
	}
}

func (f *filesystem) RemoveAll(name string) error {
	panic(100)
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
