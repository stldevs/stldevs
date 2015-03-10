package web

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockContext struct {
	calls []string

	sessionData map[string]interface{}
}

func newMockContext(init map[string]interface{}) *mockContext {
	if init != nil {
		return &mockContext{[]string{}, init}
	}
	return &mockContext{[]string{}, map[string]interface{}{}}
}

func (m *mockContext) SessionData(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	m.calls = append(m.calls, fmt.Sprint("SessionData"))
	return m.sessionData
}

func (m *mockContext) ParseAndExecute(w http.ResponseWriter, name string, data map[string]interface{}) {
	m.calls = append(m.calls, fmt.Sprintf("ParseAndExecute %v %v", name, data))
}

func TestIndex(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://www.stldevs.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	init := map[string]interface{}{"hello": "world"}
	ctx := newMockContext(init)
	index(ctx)(w, r, nil)

	if fmt.Sprintln(ctx.calls) != fmt.Sprintf("[SessionData ParseAndExecute index %v]\n", init) {
		t.Error("Result unexpected:", ctx.calls)
	}
}
