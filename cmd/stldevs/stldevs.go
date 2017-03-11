package main

import (
	"log"

	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"github.com/jakecoffman/stldevs/web"
	"github.com/jmoiron/sqlx"
	"os"
	"time"
)

const (
	TIME_BETWEEN_DB_RETRIES = time.Second
	TIME_TO_WAIT_FOR_DB = time.Minute
)


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	var db *sqlx.DB
	start := time.Now()
	first := true
	for {
		db, err = sqlx.Connect("mysql", os.Getenv("DB") + "?parseTime=true")
		if err == nil {
			break
		}
		if first {
			first = false
			log.Println("Waiting", TIME_TO_WAIT_FOR_DB, "for DB to become available")
		}
		time.Sleep(TIME_BETWEEN_DB_RETRIES)
		if time.Since(start) > TIME_TO_WAIT_FOR_DB {
			log.Fatal(err)
		}
	}
	db.MapperFunc(config.CamelToSnake)

	if err = migrations.Migrate(db); err != nil {
		log.Fatal("Could not migrate schema")
	}

	web.Run(cfg, db)
}
