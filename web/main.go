package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	fileHandler := http.FileServer(http.Dir("static/"))

	router := httprouter.New()
	router.GET("/static/*filepath", handleFiles(fileHandler))
	router.GET("/", index)
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
	http.ServeFile(w, r, "static/index.html")
}
