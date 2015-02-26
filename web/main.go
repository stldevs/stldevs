package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	fileHandler := http.FileServer(http.Dir("static/"))

	router := httprouter.New()
	router.GET("/static", fileHandler)

	log.Println("Serving on", "0.0.0.0:8070")
	http.ListenAndServe("0.0.0.0:8070", router)
}
