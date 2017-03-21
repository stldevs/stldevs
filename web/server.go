package web

import (
	"encoding/gob"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
	"gopkg.in/gin-gonic/gin.v1"
)

func Run(cfg *config.Config, db *sqlx.DB) {
	conf := &oauth2.Config{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		Scopes:       []string{},
		Endpoint:     oa2gh.Endpoint,
	}

	// for session storing
	gob.Register(github.User{})

	router := gin.Default()

	router.Use(func (c *gin.Context) {
		c.Set("db", db)
		c.Set("oauth", conf)
		c.Next()
	})

	router.GET("/search", search)
	router.GET("/toplangs", topLangs)
	router.GET("/topdevs", topDevs)
	router.GET("/toporgs", topOrgs)
	router.GET("/lang/:lang", language)
	router.GET("/profile/:profile", profile)

	log.Println("Serving on port 8080")
	log.Println(http.ListenAndServe("0.0.0.0:8080", router))
}
