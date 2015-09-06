package web

import (
	"log"
	"net/http"

	"text/template"

	"encoding/gob"

	"github.com/google/go-github/github"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jakecoffman/stldevs/aggregator"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
)

const (
	base = "web"
)

func Run(cfg *config.Config, db *sqlx.DB) {
	agg := aggregator.New(db, cfg.GithubKey)
	myDb := &DB{db}
	ctx := &contextImpl{
		store:        sessions.NewFilesystemStore("", []byte(cfg.SessionSecret)),
		trackingCode: cfg.TrackingCode,
		conf: &oauth2.Config{
			ClientID:     cfg.GithubClientID,
			ClientSecret: cfg.GithubClientSecret,
			Scopes:       []string{},
			Endpoint:     oa2gh.Endpoint,
		},
	}

	// for session storing
	gob.Register(github.User{})

	router := httprouter.New()

	router.GET("/login", login(ctx))
	router.GET("/oauth2", oauth2Handler(ctx))
	router.GET("/logout", logout(ctx))

	router.GET("/admin", admin(ctx, myDb, agg))
	router.POST("/admin", adminCmd(ctx, agg))

	router.GET("/search", search(ctx, myDb))
	router.GET("/toplangs", topLangs(ctx, myDb))
	router.GET("/topdevs", topDevs(ctx, myDb))
	router.GET("/lang/:lang", language(ctx, myDb))
	router.GET("/profile/:profile", profile(ctx, myDb))
	router.POST("/add", add(ctx, agg))

	router.PanicHandler = panicHandler

	log.Println("Serving on port 80")
	log.Println(http.ListenAndServe("0.0.0.0:80", finisher(router, ctx)))
}

func panicHandler(w http.ResponseWriter, r *http.Request, d interface{}) {
	template, err := template.ParseGlob(base + "/templates/*.html")
	if err != nil {
		log.Println(err)
		return
	}

	if err = template.ExecuteTemplate(w, "error", d); err != nil {
		log.Println(err)
	}
}

func finisher(h http.Handler, ctx Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer context.Clear(r)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		session := ctx.SessionData(w, r)
		if session != nil {
			session.Save(r, w)
			user, _ := session.Values["user"]
			if user != nil {
				log.Println(path, r.Method, r.RemoteAddr, *user.(github.User).Login)
				return
			}
		}
		log.Println(path, r.Method, r.RemoteAddr)
	})
}
