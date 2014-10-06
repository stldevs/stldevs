package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
)

func init() {
	flag.Parse()
}

type context struct {
}

type appHandler struct {
	*context
	handler func(*context, http.ResponseWriter, *http.Request) (int, interface{})
}

func (t appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, data := t.handler(t.context, w, r)
	if data == nil {
		log.Println(r.URL, "-", r.Method, "-", code, r.RemoteAddr)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("Failed to write data:", err)
	}
	log.Println(r.URL, "-", r.Method, "-", code, r.RemoteAddr)
}

func router(appCtx *context) *mux.Router {
	r := mux.NewRouter()
	// There are actually 2 options here: 1 - have 2 functions that handle both cases and
	// internally switch on the method or 2 - have a function for each method
	r.Handle("/users", appHandler{appCtx, List}).Methods("GET")
	return r
}

func main() {
	// handle all requests by serving a file of the same name
	fileHandler := http.FileServer(http.Dir("static/"))

	r := router(&context{})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileHandler))
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/404.html")
	})

	log.Println("Serving on", "0.0.0.0:8070")
	http.ListenAndServe("0.0.0.0:8070", r)
}

type Error struct {
	Error string `json:"error"`
}

func List(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	http.ServeFile(w, r, "../aggregator/users.json")
	return http.StatusOK, nil
}

// TODO: Only call on errors that are unrecoverable as the server goes down
func check(err error) {
	if err != nil {
		log.Println(err)
		debug.PrintStack()
		log.Fatal()
	}
}
