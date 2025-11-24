package sessions

import (
	"net/http"
	"sync"
	"time"

	"github.com/jakecoffman/stldevs/db"
)

const (
	Cookie     = "stldevs-session"
	KeySession = "session"
)

// Store is the global session store.
var Store = &SessionStore{
	store: map[string]*Entry{},
}

func GetEntry(r *http.Request) Entry {
	sess := r.Context().Value(KeySession)
	if sess != nil {
		entry := sess.(Entry)
		return entry
	}
	panic("No session found")
}

type SessionStore struct {
	sync.RWMutex
	store map[string]*Entry
}

type Entry struct {
	User    *db.StlDevsUser
	Created time.Time
}

func (s *SessionStore) Get(cookie string) (Entry, bool) {
	s.RLock()
	defer s.RUnlock()
	session, ok := s.store[cookie]
	if !ok {
		return Entry{}, false
	}
	return *session, ok
}

func (s *SessionStore) Add(user *db.StlDevsUser) string {
	s.Lock()
	defer s.Unlock()
	cookie := GenerateSessionCookie()
	s.store[cookie] = &Entry{
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
