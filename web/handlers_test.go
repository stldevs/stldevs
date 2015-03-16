package web

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
)

type mockContext struct {
	calls []string

	sessionData *sessions.Session
}

func newMockContext(init map[interface{}]interface{}) *mockContext {
	if init != nil {
		return &mockContext{[]string{}, &sessions.Session{Values: init}}
	}
	return &mockContext{[]string{}, &sessions.Session{}}
}

func (m *mockContext) SessionData(w http.ResponseWriter, r *http.Request) *sessions.Session {
	m.calls = append(m.calls, "SessionData")
	return m.sessionData
}

func (m *mockContext) Save(http.ResponseWriter, *http.Request) {
	m.calls = append(m.calls, "Save")
}

func (m *mockContext) ParseAndExecute(w http.ResponseWriter, name string, data map[interface{}]interface{}) {
	m.calls = append(m.calls, fmt.Sprintf("ParseAndExecute %v %v", name, data))
}

func (m *mockContext) GithubLogin(code string) (*github.User, error) {
	m.calls = append(m.calls, fmt.Sprint("GithubLogin ", code))
	return nil, nil
}

func TestIndex(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://www.stldevs.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx := newMockContext(map[interface{}]interface{}{"Hello": "world!"})
	index(ctx)(w, r, nil)

	if fmt.Sprintln(ctx.calls) != fmt.Sprintf("[SessionData ParseAndExecute index map[session:map[Hello:world!]]]\n") {
		t.Error("Result unexpected:", ctx.calls)
	}
}
