package main

import (
	"log"
	"os"

	"github.com/jakecoffman/stldevs/aggregator"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/web"
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

	if cfg.GithubKey != "" {
		agg := aggregator.New(db.DB(), cfg.GithubKey)
		agg.Schedule(db.LastRun)
	} else {
		log.Println("aggregator: GithubKey not set, skipping scheduled runs")
	}

	web.Run(cfg)
}
