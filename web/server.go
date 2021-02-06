package web

import (
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/sessions"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
	"log"
	"net/http"
)

func Run(cfg *config.Config) {
	r := gin.Default()

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

	success := &sessions.Issuer{}
	r.GET("/login", gin.WrapH(github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil))))
	r.GET("/callback", gin.WrapH(github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, success, nil))))
	r.GET("/logout", func(c *gin.Context) {
		cookie, err := c.Cookie(sessions.Cookie)
		if err != nil {
			c.JSON(200, "already logged out")
			return
		}
		sessions.Store.Evict(cookie)
		c.SetCookie(sessions.Cookie, "", -1, "/", "stldevs.com", true, true)
		c.JSON(200, "logged out")
	})

	authenticated := r.Group("", func(c *gin.Context) {
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
	})

	authenticated.GET("/me", func(c *gin.Context) {
		c.JSON(200, sessions.GetEntry(c).User)
	})

	r.GET("/search", search)

	r.GET("/last-run", func(c *gin.Context) {
		c.JSON(200, db.LastRun())
	})

	{
		devs := DevController{}
		r.GET("/devs", devs.List)
		r.GET("/devs/:login", devs.Get)
		authenticated.PATCH("/devs/:login", devs.Patch)
		authenticated.DELETE("/devs/:login", devs.Delete)
	}

	{
		orgs := OrgController{}
		r.GET("/orgs", orgs.List)
		r.GET("/orgs/:login", orgs.Get)
	}

	r.GET("/lang/:lang", language)

	// deprecated
	r.GET("/toplangs", topLangs)
	r.GET("/toporgs", topOrgs)

	log.Println("Serving on http://127.0.0.1:8080")
	log.Println(http.ListenAndServe("0.0.0.0:8080", r))
}
