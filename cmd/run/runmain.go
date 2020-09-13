package main

import (
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jakecoffman/stldevs/aggregator"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"github.com/jmoiron/sqlx"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	f, err := os.Open("./config.json")
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.NewConfig(f)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Connect("pgx", "postgres://postgres:"+cfg.PostgresPw+"@localhost:5432/stldevs")
	if err != nil {
		log.Fatal(err)
	}
	db.MapperFunc(config.CamelToSnake)

	log.Println("MIGRATE")
	if err = migrations.Migrate(db); err != nil {
		log.Fatal("Could not migrate schema")
	}

	agg := aggregator.New(db, cfg.GithubKey)
	agg.Run()
}
