package dev

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v52/github"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/sessions"
)

func TestList(t *testing.T) {
	var called bool
	db.PopularDevs = func(devType, company string) []db.DevCount {
		called = true
		if devType != "User" {
			t.Error()
		}
		return []db.DevCount{}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com?type=User", nil)
	List(w, r)

	if !called {
		t.Error()
	}
	if w.Result().StatusCode != 200 {
		t.Error(w.Result().StatusCode)
	}
}

func TestListFailure(t *testing.T) {
	var called bool
	db.PopularDevs = func(devType, company string) []db.DevCount {
		called = true
		return nil
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com?type=User", nil)
	List(w, r)

	if !called {
		t.Error()
	}
	if w.Result().StatusCode != 500 {
		t.Error(w.Result().StatusCode)
	}
}

func TestSearch(t *testing.T) {
	var called bool
	db.SearchUsers = func(term string) []db.StlDevsUser {
		called = true
		if term != "term" {
			t.Error(term)
		}
		return []db.StlDevsUser{}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com?q=term", nil)
	List(w, r)

	if !called {
		t.Error()
	}
	if w.Result().StatusCode != 200 {
		t.Error(w.Result().StatusCode)
	}
}

func TestGet(t *testing.T) {
	var called bool
	db.Profile = func(name string) (*db.ProfileData, error) {
		called = true
		if name != "bob" {
			t.Error(name)
		}
		return &db.ProfileData{}, nil
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com", nil)
	r.SetPathValue("login", "bob")
	Get(w, r)

	if !called {
		t.Error()
	}
	if w.Result().StatusCode != 200 {
		t.Error(w.Result().StatusCode)
	}
}

func TestGet404(t *testing.T) {
	var called bool
	db.Profile = func(name string) (*db.ProfileData, error) {
		called = true
		return nil, fmt.Errorf("")
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com", nil)
	r.SetPathValue("login", "bob")
	Get(w, r)

	if !called {
		t.Error()
	}
	if w.Result().StatusCode != 404 {
		t.Error(w.Result().StatusCode)
	}
}

func TestPatchByUser(t *testing.T) {
	login := "bob"
	user := &db.StlDevsUser{
		User: &github.User{Login: &login},
	}

	var called int
	db.Profile = func(name string) (*db.ProfileData, error) {
		called++
		if name != "bob" {
			t.Error()
		}
		return &db.ProfileData{User: user}, nil
	}
	db.HideUser = func(hide bool, login string) error {
		called++
		if hide != true && login != "bob" {
			t.Error(hide, login)
		}
		return nil
	}

	w := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"hide":true}`)
	r := httptest.NewRequest("GET", "http://example.com", buf)
	r.SetPathValue("login", "bob")
	ctx := context.WithValue(r.Context(), sessions.KeySession, sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	r = r.WithContext(ctx)
	Patch(w, r)

	if called != 2 {
		t.Error()
	}
	if w.Result().StatusCode != 200 {
		t.Error(w.Result().StatusCode)
	}
}

func TestPatch403(t *testing.T) {
	login := "bob"
	user := &db.StlDevsUser{
		User: &github.User{Login: &login},
	}

	w := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"hide":true}`)
	r := httptest.NewRequest("GET", "http://example.com", buf)
	r.SetPathValue("login", "alice") // bob != alice
	ctx := context.WithValue(r.Context(), sessions.KeySession, sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	r = r.WithContext(ctx)
	Patch(w, r)

	if w.Result().StatusCode != 403 {
		t.Error(w.Result().StatusCode)
	}
}

func TestPatchAdmin404(t *testing.T) {
	db.Profile = func(name string) (*db.ProfileData, error) {
		if name != "alice" {
			t.Error()
		}
		return nil, fmt.Errorf("")
	}

	login := "bob"
	user := &db.StlDevsUser{
		User:    &github.User{Login: &login},
		IsAdmin: true,
	}

	w := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"hide":true}`)
	r := httptest.NewRequest("GET", "http://example.com", buf)
	r.SetPathValue("login", "alice") // bob != alice
	ctx := context.WithValue(r.Context(), sessions.KeySession, sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	r = r.WithContext(ctx)
	Patch(w, r)

	if w.Result().StatusCode != 404 {
		t.Error(w.Result().StatusCode)
	}
}

func TestDelete(t *testing.T) {
	login := "bob"
	user := &db.StlDevsUser{
		User:    &github.User{Login: &login},
		IsAdmin: true,
	}

	db.Delete = func(login string) error {
		if login != "alice" {
			t.Errorf("%s", login)
		}
		return nil
	}

	w := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{}`)
	r := httptest.NewRequest("GET", "http://example.com", buf)
	r.SetPathValue("login", "alice") // bob != alice
	ctx := context.WithValue(r.Context(), sessions.KeySession, sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	r = r.WithContext(ctx)
	Delete(w, r)

	if w.Result().StatusCode != 200 {
		t.Error(w.Result().StatusCode)
	}
}

func TestDeleteAccessDenied(t *testing.T) {
	login := "bob"
	user := &db.StlDevsUser{
		User:    &github.User{Login: &login},
		IsAdmin: false,
	}

	db.Delete = func(login string) error {
		if login != "alice" {
			t.Errorf("%s", login)
		}
		return nil
	}

	w := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{}`)
	r := httptest.NewRequest("GET", "http://example.com", buf)
	r.SetPathValue("login", "alice") // bob != alice
	ctx := context.WithValue(r.Context(), sessions.KeySession, sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	r = r.WithContext(ctx)
	Delete(w, r)

	if w.Result().StatusCode != 403 {
		t.Error(w.Result().StatusCode)
	}
}
