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

func Run(cfg *config.Config, db *sqlx.DB, agg *aggregator.Aggregator) {
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

	fileHandler := http.FileServer(http.Dir(base + "/static/"))

	router := httprouter.New()
	router.GET("/", index(ctx))
	router.GET("/static/*filepath", handleFiles(fileHandler))

	router.GET("/login", login(ctx))
	router.GET("/oauth2", oauth2Handler(ctx))
	router.GET("/logout", logout(ctx))

	router.GET("/admin", admin(ctx, agg))
	router.POST("/admin", adminCmd(ctx, agg))

	router.GET("/search", search(ctx, agg))
	router.GET("/toplangs", topLangs(ctx, agg))
	router.GET("/topdevs", topDevs(ctx, agg))
	router.GET("/lang/:lang", language(ctx, agg))
	router.GET("/profile/:profile", profile(ctx, agg))
	router.POST("/add", add(ctx, agg))

	router.NotFound = http.HandlerFunc(notFound)
	router.PanicHandler = panicHandler

	log.Println("Serving on port 80")
	log.Println(http.ListenAndServe("0.0.0.0:80", finisher(router, ctx)))
}

func handleFiles(fileServer http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.URL.Path = p.ByName("filepath")
		fileServer.ServeHTTP(w, r)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, base+"/static/404.html")
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
		h.ServeHTTP(w, r)
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		session := ctx.SessionData(w, r)
		session.Save(r, w)
		user, _ := session.Values["user"]
		if user != nil {
			log.Println(path, r.Method, r.RemoteAddr, *user.(github.User).Login)
		} else {
			log.Println(path, r.Method, r.RemoteAddr)
		}
	})
}
