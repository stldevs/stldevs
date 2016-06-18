package web

import (
	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type Context interface {
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
