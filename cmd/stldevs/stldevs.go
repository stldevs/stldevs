package main

import (
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"github.com/jakecoffman/stldevs/web"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"time"
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

	var db *sqlx.DB
	start := time.Now()
	for {
		// postgres://pgx_md5:secret@localhost:5432/pgx_test
		db, err = sqlx.Connect("pgx", "postgres://postgres:"+cfg.PostgresPw+"@localhost:5432/stldevs")
		if err != nil {
			if time.Now().Sub(start) > 11 * time.Second {
				log.Fatal(err)
			} else {
				log.Println("failed to connect to db, trying again in 5 seconds", err)
				time.Sleep(5 * time.Second)
			}
		} else {
			break
		}
	}
	db.MapperFunc(config.CamelToSnake)

	if err = migrations.Migrate(db); err != nil {
		log.Fatal("Could not migrate schema")
	}

	web.Run(cfg, db)
}
