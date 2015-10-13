package web

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type mockContext struct {
	calls []string

	sessionData *sessions.Session
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

func (m *mockContext) AuthCodeURL(string, oauth2.AuthCodeOption) string {
	return ""
}

func (m *mockContext) GithubLogin(code string) (*github.User, error) {
	m.calls = append(m.calls, fmt.Sprint("GithubLogin ", code))
	return nil, nil
}
