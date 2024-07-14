package db

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

var db *sqlx.DB

// Connect connects to the database.
func Connect(cfg *config.Config) {
	var err error
	start := time.Now()
	for {
		// postgres://pgx_md5:secret@localhost:5432/pgx_test
		db, err = sqlx.Connect("pgx", cfg.Postgres)
		if err != nil {
			if time.Now().Sub(start) > 11*time.Second {
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
}

func Migrate() {
	if err := migrations.Migrate(db); err != nil {
		log.Fatal("Could not migrate schema")
	}
}
