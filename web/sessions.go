package web

import (
	"sync"
	"time"
)

const (
	cookieName      = "stldevs-session"
)

type SessionStore struct {
	sync.RWMutex
	store map[string]*SessionEntry
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		store: map[string]*SessionEntry{},
	}
}

type SessionEntry struct {
	User *StlDevsUser
	Created time.Time
}

func (s *SessionStore) Get(cookie string) (*SessionEntry, bool) {
	s.RLock()
	defer s.RUnlock()
	session, ok := s.store[cookie]
	return session, ok
}

func (s *SessionStore) Add(user *StlDevsUser) string {
	s.Lock()
	defer s.Unlock()
	for k, v := range s.store {
		if *v.User.ID == *user.ID {
			delete(s.store, k)
		}
	}
	cookie := GenerateSessionCookie()
	s.store[cookie] = &SessionEntry{
		User:    user,
		Created: time.Time{},
	}
	return cookie
}

func (s *SessionStore) Evict(cookie string) {
	s.Lock()
	defer s.Unlock()
	delete(s.store, cookie)
}
