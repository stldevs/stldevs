package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/db/sqlc"
	"github.com/jakecoffman/stldevs/migrations"
)

var db *sql.DB
var queries *sqlc.Queries

// Connect connects to the database.
func Connect(cfg *config.Config) {
	var err error
	start := time.Now()
	for {
		var opened *sql.DB
		opened, err = sql.Open("pgx", cfg.Postgres)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = opened.PingContext(ctx)
			cancel()
			if err == nil {
				db = opened
				break
			}
			opened.Close()
		}
		if time.Since(start) > 11*time.Second {
			log.Fatal(err)
		}
		log.Println("failed to connect to db, trying again in 5 seconds", err)
		time.Sleep(5 * time.Second)
	}
	queries = sqlc.New(db)
}

func Migrate() {
	if err := migrations.Migrate(db); err != nil {
		log.Fatal("Could not migrate schema")
	}
}
