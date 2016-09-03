package mipfs

import (
	"github.com/kpmy/mipfs/ipfs_api"
	"os"
	"path/filepath"
)

var rootName = string([]rune{os.PathSeparator})

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

func trav(root string, name string) (*loc, *link) {
	if name == rootName {
		ls, _ := ipfs_api.Shell().FileList(root)
		return &loc{*ls}, nil
	} else {
		_, last := filepath.Split(name)
		l, _ := trav(root, filepath.Dir(name))
		if li := find(l, last); li != nil {
			return li, nil
		} else {
			return l, find2(l, last)
		}
	}
}

func split(rootHash string, path string) (ret []*loc) {
	var tr func(root string) *loc
	tr = func(root string) *loc {
		if root == rootName {
			ls, _ := ipfs_api.Shell().FileList(rootHash)
			l := &loc{*ls}
			ret = append(ret, l)
			return l
		} else {
			_, last := filepath.Split(root)
			l := tr(filepath.Dir(root))
			if l != nil {
				if li := find(l, last); li != nil {
					ret = append(ret, li)
					return li
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}
	tr(path)
	return
}
