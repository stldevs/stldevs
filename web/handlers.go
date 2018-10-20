package web

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"strconv"
)

func topLangs(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	c.JSON(200, map[string]interface{}{
		"langs":   PopularLanguages(db),
		"lastrun": LastRun(db),
	})
}

func topDevs(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	c.JSON(200, map[string]interface{}{
		"devs":    PopularDevs(db),
		"lastrun": LastRun(db),
	})
}

func topOrgs(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	c.JSON(200, map[string]interface{}{
		"devs":    PopularOrgs(db),
		"lastrun": LastRun(db),
	})
}

func profile(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	profile, _ := Profile(db, c.Params.ByName("profile"))
	c.JSON(200, map[string]interface{}{
		"profile": profile,
	})
}

func language(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	pageParam := c.Request.URL.Query().Get("page")
	page := 0
	if pageParam != "" {
		var err error
		page, err = strconv.Atoi(pageParam)
		if err != nil {
			c.Status(400)
			return
		}
	}

	langs, userCount := Language(db, c.Params.ByName("lang"), page)
	c.JSON(200, map[string]interface{}{
		"languages": langs,
		"count":     userCount,
		"language":  c.Params.ByName("lang"),
		"page":      page,
	})
}

func search(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	q := c.Request.URL.Query().Get("q")
	kind := c.Request.URL.Query().Get("type")

	if q == "" {
		c.Status(400)
		return
	}

	c.JSON(200, map[string]interface{}{
		"results": Search(db, q, kind),
	})
}

func startAuth(c *gin.Context) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
		c.JSON(500, map[string]string{"error": "failed to be random"})
		return
	}
	state := base64.URLEncoding.EncodeToString(b)
	c.SetCookie("state", state, 60, "/", "", true, true)

	conf := c.MustGet("oauth").(*oauth2.Config)
	url := conf.AuthCodeURL(state)
	http.Redirect(c.Writer, c.Request, url, 302)
}

func authCallback(c *gin.Context) {
	cookie, err := c.Cookie("state")
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "are cookies disabled?"})
		return
	}
	state := c.Query("state")
	if !(len(state) != 24 && len(cookie) != 24 && state == cookie) {
		log.Println("states don't match:", state, cookie)
		c.JSON(400, gin.H{"error": "states don't match"})
		return
	}

	conf := c.MustGet("oauth").(*oauth2.Config)

	token, err := conf.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "unable to complete login"})
		return
	}

	if !token.Valid() {
		log.Println("token is invalid?")
		c.JSON(400, gin.H{"error": "token is invalid"})
		return
	}

	// use the token to get their user info
	client := github.NewClient(conf.Client(context.Background(), token))

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "unable to fetch user from github"})
		return
	}

	users := c.MustGet("users").(*UserCache)
	users.Put(token.AccessToken, token)

	c.SetCookie("access-token", token.AccessToken, 36000, "/", "", true, true)

	c.JSON(200, user)
}

func auth(c *gin.Context) {
	users := c.MustGet("users").(*UserCache)
	accessToken, err := c.Cookie("access-token")
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "missing access-token"})
		return
	}
	token, ok := users.Get(accessToken)
	if !ok {
		c.JSON(401, gin.H{"error": "no login"})
		return
	}
	if !token.Valid() {
		c.JSON(403, gin.H{"error": "token not valid, try logging in again"})
		return
	}

	c.Next()
}

func addMyself(c *gin.Context) {
	conf := c.MustGet("oauth").(*oauth2.Config)
	accessToken, err := c.Cookie("access-token")
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "no access-token, log in first?"})
		return
	}
	token := &oauth2.Token{AccessToken: accessToken}
	client := github.NewClient(conf.Client(context.Background(), token))
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "unable to fetch user from github"})
		return
	}

	// TODO
	c.JSON(200, user)
}

func removeMyself(c *gin.Context) {
	conf := c.MustGet("oauth").(*oauth2.Config)
	accessToken, err := c.Cookie("access-token")
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "no access-token, log in first?"})
		return
	}
	token := &oauth2.Token{AccessToken: accessToken}
	client := github.NewClient(conf.Client(context.Background(), token))
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "unable to fetch user from github"})
		return
	}

	// TODO
	c.JSON(200, user)
}
