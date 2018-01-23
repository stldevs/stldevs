package main

import (
	"log"
	"os"

	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"github.com/jakecoffman/stldevs/web"
	"github.com/jmoiron/sqlx"
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
		db, err = sqlx.Connect("mysql", "root:"+cfg.MysqlPw+"@/stldevs?parseTime=true")
		if err != nil {
			if time.Now().Sub(start) > 11 * time.Second {
				log.Fatal(err)
			} else {
				log.Println("failed to connect to mysql, trying again in 5 seconds")
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
