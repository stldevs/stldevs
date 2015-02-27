package web

import (
	"log"
	"net/http"

	"text/template"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
)

const (
	base = "web"
)

var store = sessions.NewFilesystemStore("", []byte("secret")) // TODO

func Run() {
	fileHandler := http.FileServer(http.Dir(base + "/static/"))

	router := httprouter.New()
	router.GET("/static/*filepath", handleFiles(fileHandler))
	router.GET("/", index)
	router.GET("/oauth2", oauth2Handler)
	router.GET("/logout", logout)
	router.NotFound = http.HandlerFunc(notFound)
	router.PanicHandler = panicHandler

	log.Println("Serving on", "localhost:80")
	log.Println(http.ListenAndServe("localhost:80", context.ClearHandler(router)))
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

func panicHandler(w http.ResponseWriter, r *http.Request, _ interface{}) {
	w.WriteHeader(500)
	http.ServeFile(w, r, base+"/static/500.html")
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	template, err := template.ParseGlob(base + "/templates/*.html")
	if err != nil {
		log.Println(err)
		return
	}

	data := map[string]string{}
	user, _ := get_session(r, "user")
	if user != "" {
		data["user"] = user
	} else {
		set_session(w, r, "githubState", randSeq(10))
		data["github"] = conf.AuthCodeURL("statey", oauth2.AccessTypeOffline)
	}

	if err = template.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Println(err)
		return
	}
}
