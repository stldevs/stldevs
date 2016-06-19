package web

import (
	"encoding/gob"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
)

func Run(cfg *config.Config, db *sqlx.DB) {
	services := &Stldevs{
		db,
		&oauth2.Config{
			ClientID:     cfg.GithubClientID,
			ClientSecret: cfg.GithubClientSecret,
			Scopes:       []string{},
			Endpoint:     oa2gh.Endpoint,
		},
	}

	// for session storing
	gob.Register(github.User{})

	router := httprouter.New()

	router.GET("/search", mw(services, search))
	router.GET("/toplangs", mw(services, topLangs))
	router.GET("/topdevs", mw(services, topDevs))
	router.GET("/toporgs", mw(services, topOrgs))
	router.GET("/lang/:lang", mw(services, language))
	router.GET("/profile/:profile", mw(services, profile))

	router.PanicHandler = panicHandler

	log.Println("Serving on port 8080")
	log.Println(http.ListenAndServe("0.0.0.0:8080", router))
}
