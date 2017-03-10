package main

import (
	"log"

	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"github.com/jakecoffman/stldevs/web"
	"github.com/jmoiron/sqlx"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Connect("mysql", os.Getenv("DB") + "?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	db.MapperFunc(config.CamelToSnake)

	if err = migrations.Migrate(db); err != nil {
		log.Fatal("Could not migrate schema")
	}

	web.Run(cfg, db)
}
