package wdfs

import (
	"bytes"
	"encoding/xml"
	"github.com/ipfs/go-ipfs-api"
	"github.com/kpmy/ypk/dom"
	"github.com/kpmy/ypk/fn"
	. "github.com/kpmy/ypk/tc"
	"golang.org/x/net/webdav"
	"io"
	"reflect"
	"strings"
)

func newPropsModel() (ret dom.Element) {
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
	for k := range model.AttrAsMap() {
		xn := xml.Name{Local: k, Space: "IPFSATTR"}
		ret[xn] = model
	}
	return
}

func writeProps(props map[xml.Name]dom.Element) (ret dom.Element) {
	ret = dom.Elem("props")
	for k, v := range props {
		if k.Space == "IPFSATTR" {
			ret.Attr(k.Local, v.Attr(k.Local))
		} else {
			ret.AppendChild(v)
		}
	}
	return
}

func props2webdav(pm map[xml.Name]dom.Element) (ret map[xml.Name]webdav.Property) {
	ret = make(map[xml.Name]webdav.Property)
	for k, v := range pm {
		p := webdav.Property{XMLName: k}
		buf := new(bytes.Buffer)
		if k.Space == "IPFSATTR" {
			buf.WriteString(v.Attr(k.Local))
		} else {
			Assert(v.ChildrenCount() == 1, 40)
			c0 := v.Children()[0]
			switch c := c0.(type) {
			case dom.Element:
				rd := dom.Encode(c)
				io.Copy(buf, rd)
			case dom.Text:
				xml.EscapeText(buf, []byte(c.Data()))
			default:
				Halt(100, reflect.TypeOf(c))
			}
		}
		p.InnerXML = buf.Bytes()
		ret[k] = p
	}
	return
}

func propsPatch(pe map[xml.Name]dom.Element, patch []webdav.Proppatch) (ret []webdav.Propstat) {
	ret = []webdav.Propstat{}
	for _, pl := range patch {
		ps := webdav.Propstat{}
		for _, p := range pl.Props {
			if pl.Remove {
				delete(pe, p.XMLName)
			} else if p.XMLName.Space == "IPFSATTR" {
				tmp := dom.Elem("props")
				tmp.Attr(p.XMLName.Local, string(p.InnerXML))
				pe[p.XMLName] = tmp
			} else {
				el := dom.Elem("prop")
				el.Attr("local", p.XMLName.Local)
				el.Attr("space", p.XMLName.Space)
				e, _ := dom.Decode(bytes.NewBuffer(p.InnerXML))
				if !fn.IsNil(e.Model()) {
					el.AppendChild(e.Model())
				} else if !fn.IsNil(e.Data()) {
					el.AppendChild(e.Data())
				} else {
					Halt(100)
				}
				pe[p.XMLName] = el
			}
			ps.Props = append(ps.Props, p)
		}
		ps.Status = 200
		ret = append(ret, ps)
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
