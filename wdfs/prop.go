package wdfs

import (
	"github.com/ipfs/go-ipfs-api"
	"github.com/kpmy/ypk/dom"
	"strings"
)

func newProps() (ret dom.Element) {
	ret = dom.Elem("props")
	return
}

func propsMap(obj *shell.UnixLsObject) (ret map[string]*shell.UnixLsLink) {
	ret = make(map[string]*shell.UnixLsLink)
	for _, lo := range obj.Links {
		if lo.Type == "File" {
			if lo.Name == "*" {
				ret["*"] = lo
			} else if strings.HasPrefix(lo.Name, "*") {
				ret[strings.Trim(lo.Name, "*")] = lo
			}
		}
	}
	return
}
