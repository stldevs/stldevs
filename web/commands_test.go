package web

import (
	"io"
	"testing"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func init() {
	var err error
	for {
		db, err = sqlx.Connect("pgx", "postgres://postgres:pw@127.0.0.1:5432/postgres")
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
		} else {
			break
		}
	}
	db.MapperFunc(config.CamelToSnake)
	db.MustExec("select 1")
}

func TestMigrate(t *testing.T) {
	if err := migrations.Migrate(db); err != nil {
		t.Error(err)
	}
}

func TestLastRun(t *testing.T) {
	if v := LastRun(db); v == nil || !v.Equal(time.Time{}) {
		t.Errorf("Time should have been zero value, got %v", v)
	}
	db.MustExec("insert into agg_meta values (CURRENT_TIMESTAMP)")
	if v := LastRun(db); v == nil || !v.After(time.Time{}) {
		t.Errorf("Time should have been greater than zero value, got %v", v)
	}
}
