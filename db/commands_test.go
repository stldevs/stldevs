package db

import (
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
	"log"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	Connect(&config.Config{
		Postgres: "postgres://postgres:pw@127.0.0.1:5432/postgres",
	})
	db.MustExec("drop table agg_meta")
	db.MustExec("drop table agg_repo")
	db.MustExec("drop table agg_user")
	db.MustExec("drop table migrations")
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

func TestHideUser(t *testing.T) {
	db.MustExec("insert into agg_user (login, hide) values ('bob', false) on conflict do nothing")
	if err := HideUser(true, "bob"); err != nil {
		t.Fatal(err)
	}
	user, err := Profile("bob")
	if err != nil {
		t.Fatal(err)
	}
	if !user.User.Hide {
		t.Fatal("expected hidden, was not")
	}

	if err = HideUser(false, "bob"); err != nil {
		t.Fatal(err)
	}
	user, err = Profile("bob")
	if err != nil {
		t.Fatal(err)
	}
	if user.User.Hide {
		t.Fatal("expected shown, was not")
	}
}

func TestPopularDevs(t *testing.T) {
	result := PopularDevs("Butt")
	if len(result) != 0 {
		t.Error(len(result))
	}
}
