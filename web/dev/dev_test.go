package dev

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/sessions"
	"net/http/httptest"
	"testing"
	"time"
)

func TestList(t *testing.T) {
	var called bool
	db.PopularDevs = func() []db.DevCount {
		called = true
		return []db.DevCount{}
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://example.com", nil)
	List(c)

	if !called {
		t.Error()
	}
	if w.Result().StatusCode != 200 {
		t.Error(w.Result().StatusCode)
	}
}

func TestListFailure(t *testing.T) {
	var called bool
	db.PopularDevs = func() []db.DevCount {
		called = true
		return nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://example.com", nil)
	List(c)

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
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://example.com?q=term", nil)
	List(c)

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
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://example.com", nil)
	c.Params = gin.Params{{Key: "login", Value: "bob"}}
	Get(c)

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
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://example.com", nil)
	c.Params = gin.Params{{Key: "login", Value: "bob"}}
	Get(c)

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
	c, _ := gin.CreateTestContext(w)
	buf := bytes.NewBufferString(`{"hide":true}`)
	c.Request = httptest.NewRequest("GET", "http://example.com", buf)
	c.Params = gin.Params{{Key: "login", Value: "bob"}}
	c.Set(sessions.KeySession, &sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	Patch(c)

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
	c, _ := gin.CreateTestContext(w)
	buf := bytes.NewBufferString(`{"hide":true}`)
	c.Request = httptest.NewRequest("GET", "http://example.com", buf)
	c.Params = gin.Params{{Key: "login", Value: "alice"}} // bob != alice
	c.Set(sessions.KeySession, &sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	Patch(c)

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
	c, _ := gin.CreateTestContext(w)
	buf := bytes.NewBufferString(`{"hide":true}`)
	c.Request = httptest.NewRequest("GET", "http://example.com", buf)
	c.Params = gin.Params{{Key: "login", Value: "alice"}} // bob != alice
	c.Set(sessions.KeySession, &sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	Patch(c)

	if w.Result().StatusCode != 404 {
		t.Error(w.Result().StatusCode)
	}
}

func TestPatchBindFailed(t *testing.T) {
	login := "bob"
	user := &db.StlDevsUser{
		User:    &github.User{Login: &login},
		IsAdmin: true,
	}

	db.Profile = func(name string) (*db.ProfileData, error) {
		if name != "alice" {
			t.Error()
		}
		return &db.ProfileData{
			User: user,
		}, nil
	}
	db.HideUser = func(hide bool, login string) error {
		t.Errorf("should not have got here")
		return fmt.Errorf("")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	buf := bytes.NewBufferString(`{}`)
	c.Request = httptest.NewRequest("GET", "http://example.com", buf)
	c.Params = gin.Params{{Key: "login", Value: "alice"}} // bob != alice
	c.Set(sessions.KeySession, &sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	Patch(c)

	if w.Result().StatusCode != 400 {
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
			t.Errorf(login)
		}
		return nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	buf := bytes.NewBufferString(`{}`)
	c.Request = httptest.NewRequest("GET", "http://example.com", buf)
	c.Params = gin.Params{{Key: "login", Value: "alice"}} // bob != alice
	c.Set(sessions.KeySession, &sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	Delete(c)

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
			t.Errorf(login)
		}
		return nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	buf := bytes.NewBufferString(`{}`)
	c.Request = httptest.NewRequest("GET", "http://example.com", buf)
	c.Params = gin.Params{{Key: "login", Value: "alice"}} // bob != alice
	c.Set(sessions.KeySession, &sessions.Entry{
		User:    user,
		Created: time.Now(),
	})
	Delete(c)

	if w.Result().StatusCode != 403 {
		t.Error(w.Result().StatusCode)
	}
}
