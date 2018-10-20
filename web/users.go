package web

import (
	"golang.org/x/oauth2"
	"sync"
)

type UserCache struct {
	sync.RWMutex
	lookup map[string]*oauth2.Token
}

func NewUserCache() *UserCache {
	return &UserCache{lookup: map[string]*oauth2.Token{}}
}

func (u *UserCache) Get(session string) (*oauth2.Token, bool) {
	u.RLock()
	token, ok := u.lookup[session]
	u.RUnlock()
	return token, ok
}

func (u *UserCache) Put(session string, token *oauth2.Token) {
	u.Lock()
	u.lookup[session] = token
	u.Unlock()
}
