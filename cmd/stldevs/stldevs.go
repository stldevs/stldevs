package main

import (
	"log"
	"os"

	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/web"
	"github.com/jmoiron/sqlx"
	"github.com/jakecoffman/stldevs/migrations"
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

	db, err := sqlx.Connect("mysql", "root:"+cfg.MysqlPw+"@/stldevs?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	db.MapperFunc(config.CamelToSnake)

	if err = migrations.Migrate(db); err != nil {
		log.Fatal("Could not migrate schema")
	}

	web.Run(cfg, db)
}
