package mipfs

import (
	"github.com/kpmy/mipfs/ipfs_api"
	"os"
	"path/filepath"
)

var root = string([]rune{os.PathSeparator})

func find(li *loc, name string) (ret *loc) {
	for _, i := range li.Links {
		if i.Type == "Directory" && i.Name == name {
			ls, _ := ipfs_api.Shell().FileList(i.Hash)
			return &loc{*ls}
		}
	}
	return
}

func find2(li *loc, name string) (ret *link) {
	for _, i := range li.Links {
		if i.Type == "File" && i.Name == name {
			return &link{*i}
		}
	}
	return
}

func trav(rootHash string, name string) (*loc, *link) {
	if name == root {
		ls, _ := ipfs_api.Shell().FileList(rootHash)
		return &loc{*ls}, nil
	} else {
		_, last := filepath.Split(name)
		l, _ := trav(rootHash, filepath.Dir(name))
		if li := find(l, last); li != nil {
			return li, nil
		} else {
			return l, find2(l, last)
		}
	}
}
