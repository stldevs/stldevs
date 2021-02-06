package web

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/jakecoffman/stldevs/db"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/config"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
)

func Run(cfg *config.Config) {
	sessionStore := NewSessionStore()

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

	success := &sessionIssuer{store: sessionStore}
	r.GET("/login", gin.WrapH(github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil))))
	r.GET("/callback", gin.WrapH(github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, success, nil))))
	r.GET("/logout", func(c *gin.Context) {
		cookie, err := c.Cookie(cookieName)
		if err != nil {
			c.JSON(200, "already logged out")
			return
		}
		sessionStore.Evict(cookie)
		c.SetCookie(cookieName, "", -1, "/", "stldevs.com", true, true)
		c.JSON(200, "logged out")
	})

	authenticated := r.Group("", func(c *gin.Context) {
		cookie, err := c.Cookie(cookieName)
		if err != nil || cookie == "" {
			c.AbortWithStatusJSON(401, "Not logged in")
			return
		}
		session, ok := sessionStore.Get(cookie)
		if !ok {
			c.AbortWithStatusJSON(401, "Not logged in")
			return
		}
		c.Set("session", session)
	})

	authenticated.GET("/me", func(c *gin.Context) {
		session, _ := c.Get("session")
		c.JSON(200, session.(*SessionEntry).User)
	})

	r.GET("/search", search)

	r.GET("/last-run", func(c *gin.Context) {
		c.JSON(200, db.LastRun())
	})

	{
		devs := DevController{store: sessionStore}
		r.GET("/devs", devs.List)
		r.GET("/devs/:login", devs.Get)
		authenticated.PATCH("/devs/:login", devs.Patch)
		authenticated.DELETE("/devs/:login", devs.Delete)
	}

	{
		orgs := OrgController{store: sessionStore}
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

type sessionIssuer struct {
	store *SessionStore
}

func (s *sessionIssuer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	githubUser, err := github.UserFromContext(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		_, _ = w.Write([]byte("\"" + err.Error() + "\""))
		return
	}

	log.Println("Login success", *githubUser.Login)

	user := &db.StlDevsUser{
		User: githubUser,
	}

	// check if the user is an admin and set that in the session too
	user.IsAdmin = db.IsAdmin(*githubUser.Login)

	expire := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{
		Name:    cookieName,
		Value:   s.store.Add(user),
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/you", http.StatusFound)
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateSessionCookie() string {
	b, _ := GenerateRandomBytes(64)
	return base64.URLEncoding.EncodeToString(b)
}
