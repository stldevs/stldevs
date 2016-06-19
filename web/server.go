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

	router.GET("/search", search(services))
	router.GET("/toplangs", topLangs(services))
	router.GET("/topdevs", topDevs(services))
	router.GET("/toporgs", topOrgs(services))
	router.GET("/lang/:lang", language(services))
	router.GET("/profile/:profile", profile(services))

	router.PanicHandler = panicHandler

	log.Println("Serving on port 8080")
	log.Println(http.ListenAndServe("0.0.0.0:8080", finisher(router)))
}

func panicHandler(w http.ResponseWriter, _ *http.Request, d interface{}) {
	log.Println("ERROR WAS:", d)
	w.WriteHeader(500)
	w.Write([]byte("There was an error"))
}

func finisher(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		log.Println(path, r.Method, r.RemoteAddr)
	})
}
