package main

import (
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/web"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	f, err := os.Open("./config.json") // TODO: Make configurable
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.NewConfig(f)
	if err != nil {
		log.Fatal(err)
	}

	db.Connect(cfg)
	db.Migrate()
	web.Run(cfg)
}
