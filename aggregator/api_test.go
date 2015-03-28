package aggregator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jakecoffman/stldevs/config"
	"github.com/jmoiron/sqlx"
)

func setup() *sqlx.DB {
	cfg := config.Config{}
	f, err := os.Open("../config.json")
	if err != nil {
		log.Fatal("Couldn't find dev_config.json")
	}

	json.NewDecoder(f).Decode(&cfg)

	db, err := sqlx.Connect("mysql", "root:"+cfg.MysqlPw+"@/stldevs?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	db.MapperFunc(config.CamelToSnake)
	return db
}

func TestLanguage(t *testing.T) {
	db := setup()
	agg := New(db, "")

	for _, result := range agg.Language("Go") {
		fmt.Printf("%v\n", result)
	}
}
