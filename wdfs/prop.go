package wdfs

import (
	"encoding/xml"
	"github.com/ipfs/go-ipfs-api"
	"github.com/kpmy/ypk/dom"
	. "github.com/kpmy/ypk/tc"
	"reflect"
	"strings"
)

func newProps() (ret dom.Element) {
	ret = dom.Elem("props")
	return
}

func readProps(model dom.Element) (ret map[xml.Name]dom.Element) {
	ret = make(map[xml.Name]dom.Element)
	for _, _e := range model.Children() {
		switch e := _e.(type) {
		case dom.Element:
			xn := xml.Name{Local: e.Attr("local"), Space: e.Attr("space")}
			ret[xn] = e
		default:
			Halt(100, reflect.TypeOf(e))
		}
	}
	return
}

func writeProps(props map[xml.Name]dom.Element) (ret dom.Element) {
	ret = dom.Elem("props")
	for _, v := range props {
		ret.AppendChild(v)
	}
	return
}

func propLinksMap(obj *shell.UnixLsObject) (ret map[string]*shell.UnixLsLink) {
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
