package wdfs

import (
	"fmt"
	"golang.org/x/net/webdav"
	"strings"
	"sync"
	"time"
)

type locksystem struct {
	webdav.LockSystem
	fs *filesystem
	sync.RWMutex
	locks  map[string]string
	tokens map[string]webdav.LockDetails
	holds  map[string]bool
	idx    chan string
}

func (l *locksystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	const noLock = "DAV:no-lock"
	//log.Println("confirm", name0, name1, conditions)
	l.RLock()
	tok0, ok0 := l.locks[name0]
	if _, ok1 := l.locks[name1]; ok1 {
		panic(101)
	}

	ok := true
	etag, _ := l.fs.ETag(name0)
	for _, c := range conditions {
		token := c.Token
		if c.ETag != "" {
			tok0 = etag
			token = c.ETag
		}
		switch {
		case !c.Not && ok0:
			ok = tok0 == token
		case token == noLock:
			ok = !ok0 && c.Not
		default:
			ok = false
		}
		if ok == false {
			break
		}
	}
	//log.Println(ok)
	if ok && !l.holds[name0] {
		l.RUnlock()
		l.Lock()
		l.holds[name0] = true
		release = func() {
			delete(l.holds, name0)
			//log.Println(name0, "release")
		}
		l.RWMutex.Unlock()
	} else {
		err = webdav.ErrConfirmationFailed
		l.RUnlock()
	}
	return
}

func (l *locksystem) ParentLocks(name string) (ret []string) {
	l.RLock()
	ls := []string{"/"}
	ls = append(ls, strings.Split(strings.Trim(name, "/"), "/")...)
	var path = ""
	for i := 0; i < len(ls); i++ {
		path = path + ls[i]
		if tok, ok := l.locks[path]; ok {
			if details := l.tokens[tok]; !details.ZeroDepth {
				ret = append(ret, tok)
			}
		}
		if i > 0 {
			path = path + "/"
			if tok, ok := l.locks[path]; ok {
				if details := l.tokens[tok]; !details.ZeroDepth {
					ret = append(ret, tok)
				}
			}
		}
	}
	l.RUnlock()
	return
}

func (l *locksystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	//log.Println("lock", details)
	l.RLock()
	if _, ok := l.locks[details.Root]; !ok && len(l.ParentLocks(details.Root)) == 0 {
		l.RUnlock()
		l.Lock()
		token = <-l.idx + ":" + fmt.Sprint(now.UnixNano())
		l.locks[details.Root] = token
		l.tokens[token] = details
		//log.Println("locked", token)
		l.RWMutex.Unlock()
	} else {
		l.RUnlock()
		err = webdav.ErrLocked
	}
	return
}

func (l *locksystem) Refresh(now time.Time, token string, duration time.Duration) (ret webdav.LockDetails, err error) {
	l.Lock()
	if details, ok := l.tokens[token]; ok {
		details.Duration = duration
		l.tokens[token] = details
		ret = details
	} else {
		err = webdav.ErrNoSuchLock
	}
	l.RWMutex.Unlock()
	return
}

func (l *locksystem) Unlock(now time.Time, token string) (err error) {
	//log.Println("unlock", token)
	l.Lock()
	if details, ok := l.tokens[token]; ok {
		if !l.holds[details.Root] {
			delete(l.tokens, token)
			delete(l.locks, details.Root)
		} else {
			err = webdav.ErrLocked
		}
	} else {
		err = webdav.ErrNoSuchLock
	}
	l.RWMutex.Unlock()
	return
}

func NewLS(fs *filesystem) *locksystem {
	ret := &locksystem{}
	ret.fs = fs
	ret.locks = make(map[string]string)
	ret.tokens = make(map[string]webdav.LockDetails)
	ret.holds = make(map[string]bool)
	ret.idx = make(chan string)
	go func(ch chan string) {
		i := 0
		for {
			ch <- fmt.Sprint(i)
			i++
		}
	}(ret.idx)
	return ret
}
