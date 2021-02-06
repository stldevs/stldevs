package db

import (
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"testing"
	"time"
)

func init() {
	Connect(&config.Config{
		Postgres: "postgres://postgres:pw@127.0.0.1:5432/postgres",
	})
	db.MustExec("drop table agg_meta")
}

func TestMigrate(t *testing.T) {
	err := migrations.Migrate(db)
	if err != nil {
		t.Error(err)
	}
}

func TestLastRun(t *testing.T) {
	if v := LastRun(); !v.Equal(time.Time{}) {
		t.Errorf("Time should have been zero value, got %v", v)
	}
	db.MustExec("insert into agg_meta values (CURRENT_TIMESTAMP)")
	if v := LastRun(); !v.After(time.Time{}) {
		t.Errorf("Time should have been greater than zero value, got %v", v)
	}
}
