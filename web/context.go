package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type Context interface {
	// gets common session data, like user
	SessionData(http.ResponseWriter, *http.Request) *sessions.Session
	// Saves changes to session data
	Save(http.ResponseWriter, *http.Request)
	// parses and executes template
	ParseAndExecute(http.ResponseWriter, string, map[interface{}]interface{})
	// gets login URL
	AuthCodeURL(string, oauth2.AuthCodeOption) string
	// logs in with github
	GithubLogin(code string) (*github.User, error)
}

type contextImpl struct {
	store        *sessions.FilesystemStore
	trackingCode string
	conf         *oauth2.Config
}

// TODO in production we want to just parse once
func (c *contextImpl) ParseAndExecute(w http.ResponseWriter, templateName string, data map[interface{}]interface{}) {
	template, err := template.ParseGlob(base + "/templates/*.html")
	if err != nil {
		panic(err)
	}
	data["page"] = templateName
	if err = template.ExecuteTemplate(w, templateName, data); err != nil {
		panic(err)
	}
}

func (c *contextImpl) AuthCodeURL(state string, option oauth2.AuthCodeOption) string {
	return c.conf.AuthCodeURL(state, option)
}

func (c *contextImpl) GithubLogin(code string) (*github.User, error) {
	token, err := c.conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	client := github.NewClient(c.conf.Client(oauth2.NoContext, token))

	user, _, err := client.Users.Get("")
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *contextImpl) SessionData(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := c.store.Get(r, "session")
	if err != nil {
		log.Println(err)
		return nil
	}
	session.Values["trackingCode"] = c.trackingCode
	user, _ := session.Values["user"]
	if user != nil {
		// TODO extract an admin list
		if *user.(github.User).Login == "jakecoffman" {
			session.Values["admin"] = true
		}
	}
	return session
}

func (c *contextImpl) Save(w http.ResponseWriter, r *http.Request) {
	session, err := c.store.Get(r, "session")
	if err != nil {
		log.Println("In Save:", err)
		return
	}
	if err = session.Save(r, w); err != nil {
		log.Println("In Save:", err)
	}
}
