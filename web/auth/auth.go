package auth

import (
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/sessions"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
)

func New(cfg *config.Config) []crud.Spec {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint:     oa2gh.Endpoint,
	}

	var stateConfig gologin.CookieConfig
	if cfg.Environment == "prod" {
		oauth2Config.RedirectURL = "https://stldevs.com/stldevs-api/callback"
		stateConfig = gologin.CookieConfig{
			Name:     "stldevs",
			Path:     "/",
			MaxAge:   60,
			HTTPOnly: true,
			Secure:   true, // secure only
		}
	} else {
		stateConfig = gologin.DebugOnlyCookieConfig
	}

	loginTags := []string{"Login"}

	success := &sessions.Issuer{}
	return []crud.Spec{{
		Method:      "GET",
		Path:        "/login",
		Handler:     gin.WrapH(github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil))),
		Description: "GitHub OAuth Login",
		Tags:        loginTags,
	}, {
		Method:      "GET",
		Path:        "/callback",
		Handler:     gin.WrapH(github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, success, nil))),
		Description: "GitHub OAuth Callback",
		Tags:        loginTags,
	}, {
		Method:      "GET",
		Path:        "/logout",
		Handler:     logout,
		Description: "Logout of session",
		Tags:        loginTags,
	}, {
		Method:      "GET",
		Path:        "/me",
		PreHandlers: []gin.HandlerFunc{authenticated},
		Handler:     me,
		Description: "Get info about the logged in user",
		Tags:        loginTags,
	}}
}

func authenticated(c *gin.Context) {
	cookie, err := c.Cookie(sessions.Cookie)
	if err != nil || cookie == "" {
		c.AbortWithStatusJSON(401, "Not logged in")
		return
	}
	session, ok := sessions.Store.Get(cookie)
	if !ok {
		c.AbortWithStatusJSON(401, "Not logged in")
		return
	}
	c.Set(sessions.KeySession, session)
}

func me(c *gin.Context) {
	c.JSON(200, sessions.GetEntry(c).User)
}

func logout(c *gin.Context) {
	cookie, err := c.Cookie(sessions.Cookie)
	if err != nil {
		c.JSON(200, "already logged out")
		return
	}
	sessions.Store.Evict(cookie)
	c.SetCookie(sessions.Cookie, "", -1, "/", "stldevs.com", true, true)
	c.JSON(200, "logged out")
}
