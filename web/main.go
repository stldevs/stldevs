package web

import (
	"log"
	"net/http"

	"text/template"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
)

func Run() {
	fileHandler := http.FileServer(http.Dir("static/"))

	router := httprouter.New()
	router.GET("/static/*filepath", handleFiles(fileHandler))
	router.GET("/", index)
	router.GET("/oauth2", oauth2Handler)
	router.NotFound = http.HandlerFunc(notFound)

	log.Println("Serving on", "localhost:80")
	log.Println(http.ListenAndServe("localhost:80", router))
}

func handleFiles(fileServer http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.URL.Path = p.ByName("filepath")
		fileServer.ServeHTTP(w, r)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/404.html")
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	template, err := template.New("index.html").ParseFiles("templates/index.html")
	if err != nil {
		log.Println(err)
		return
	}
	data := map[string]string{}
	// TODO: Set state randomly for CSRF protection
	data["github"] = conf.AuthCodeURL("statey", oauth2.AccessTypeOffline)
	if err = template.Execute(w, data); err != nil {
		log.Println(err)
		return
	}
}
