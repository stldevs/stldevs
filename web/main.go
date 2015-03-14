package web

import (
	"log"
	"net/http"

	"text/template"

	"database/sql"

	"encoding/gob"

	"github.com/google/go-github/github"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/jakecoffman/stldevs/aggregator"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
)

const (
	base = "web"
)

var conf *oauth2.Config
var store *sessions.FilesystemStore
var trackingCode string

type Config struct {
	GithubKey, MysqlPw, GithubClientID, GithubClientSecret, SessionSecret, TrackingCode string
}

type Context interface {
	// gets common session data, like user
	SessionData(http.ResponseWriter, *http.Request) map[string]interface{}
	// parses and executes template
	ParseAndExecute(http.ResponseWriter, string, map[string]interface{})
}

func Run(config Config) {
	store = sessions.NewFilesystemStore("", []byte(config.SessionSecret))
	trackingCode = config.TrackingCode

	conf = &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		Scopes:       []string{"public_repo"},
		Endpoint:     oa2gh.Endpoint,
	}

	db, err := sql.Open("mysql", "root:"+config.MysqlPw+"@/stldevs")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	agg := aggregator.New(db, config.GithubKey)

	ctx := &contextImpl{}

	gob.Register(github.User{})

	fileHandler := http.FileServer(http.Dir(base + "/static/"))

	router := httprouter.New()
	router.GET("/static/*filepath", handleFiles(fileHandler))
	router.GET("/oauth2", oauth2Handler(ctx))
	router.GET("/logout", logout)
	router.GET("/", index(ctx))
	router.GET("/admin", admin(ctx, agg))
	router.POST("/admin", adminCmd(ctx, agg))
	router.GET("/search", search(ctx, agg))
	router.GET("/toplangs", topLangs(ctx, agg))
	router.GET("/topdevs", topDevs(ctx, agg))
	router.GET("/profile/:profile", profile(ctx, agg))
	router.POST("/add", add(ctx, agg))
	router.GET("/lang/:lang", language(ctx, agg))
	router.NotFound = http.HandlerFunc(notFound)
	router.PanicHandler = panicHandler

	log.Println("Serving on port 80")
	log.Println(http.ListenAndServe("0.0.0.0:80", context.ClearHandler(Logger(router))))
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

type contextImpl struct{}

// TODO in production we want to just parse once
func (c *contextImpl) ParseAndExecute(w http.ResponseWriter, templateName string, data map[string]interface{}) {
	template, err := template.ParseGlob(base + "/templates/*.html")
	if err != nil {
		panic(err)
	}
	data["page"] = templateName
	if err = template.ExecuteTemplate(w, templateName, data); err != nil {
		panic(err)
	}
}

func (c *contextImpl) SessionData(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	data := map[string]interface{}{}
	data["trackingCode"] = trackingCode
	user, _ := get_session(r, "user")
	if user != nil {
		data["user"] = user
		// TODO extract an admin list
		if *user.(github.User).Login == "jakecoffman" {
			data["admin"] = true
		}
	} else {
		state := randSeq(10)
		set_session(w, r, "state", state)
		data["github"] = conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	}
	return data
}

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		user, _ := get_session(r, "user")
		if user != nil {
			log.Println(path, r.Method, r.RemoteAddr, *user.(github.User).Login)
		} else {
			log.Println(path, r.Method, r.RemoteAddr)
		}
	})
}
