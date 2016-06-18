package main

import (
	"github.com/jakecoffman/stldevs/aggregator"
	"log"
	"github.com/jmoiron/sqlx"
	"github.com/jakecoffman/stldevs/config"
	"os"
)

func main() {
	f, err := os.Open("./config.json") // TODO: make configurable
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

	agg := aggregator.New(db, cfg.GithubKey)
	u, _ := agg.FindInStl("org")
	log.Println(len(u))
}

