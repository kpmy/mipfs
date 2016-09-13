package wdfs

import (
	"fmt"
	"golang.org/x/net/webdav"
	"log"
	"sync"
	"time"
)

type locksystem struct {
	webdav.LockSystem
	sync.RWMutex
	locks  map[string]string
	tokens map[string]webdav.LockDetails
	idx    chan string
}

func (l *locksystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	log.Println("confirm", name0, name1, conditions)
	l.RLock()
	if _, ok := l.locks[name0]; ok {
		release = func() {
			log.Println(name0, "release")
		}
	} else {
		err = webdav.ErrConfirmationFailed
	}
	l.RUnlock()
	return
}

func (l *locksystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	log.Println("lock", details)
	l.RLock()
	if _, ok := l.locks[details.Root]; !ok {
		l.RUnlock()
		l.Lock()
		token = <-l.idx + ":" + fmt.Sprint(now.UnixNano())
		l.locks[details.Root] = token
		l.tokens[token] = details
		log.Println("locked", token)
		l.RWMutex.Unlock()
	} else {
		l.RUnlock()
		err = webdav.ErrLocked
	}
	return
}

func (l *locksystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	panic(100)
}

func (l *locksystem) Unlock(now time.Time, token string) (err error) {
	log.Println("unlock", token)
	l.Lock()
	details := l.tokens[token]
	delete(l.tokens, token)
	delete(l.locks, details.Root)
	l.RWMutex.Unlock()
	return
}

func NewLS(fs webdav.FileSystem) *locksystem {
	ret := &locksystem{}
	ret.locks = make(map[string]string)
	ret.tokens = make(map[string]webdav.LockDetails)
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
