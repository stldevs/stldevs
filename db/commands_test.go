package db

import (
	"log"
	"testing"
	"time"

	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/migrations"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	Connect(&config.Config{
		Postgres: "postgres://postgres:pw@127.0.0.1:5432/postgres",
	})
	mustExec("drop table if exists agg_meta")
	mustExec("drop table if exists agg_repo")
	mustExec("drop table if exists agg_user")
	mustExec("drop table if exists migrations")
	Migrate()
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
	mustExec("insert into agg_meta values (CURRENT_TIMESTAMP)")
	if v := LastRun(); !v.After(time.Time{}) {
		t.Errorf("Time should have been greater than zero value, got %v", v)
	}
}

func TestHideUser(t *testing.T) {
	mustExec("insert into agg_user (login, company, hide) values ('bob', '', false) on conflict do nothing")
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
	result := PopularDevs("User", "company")
	if len(result) != 0 {
		t.Error(len(result))
	}
}

func TestPopularDevsCompanyFilter(t *testing.T) {
	resetTables(t)
	const (
		login   = "popular-dev"
		company = "Acme Corp"
	)
	mustExec(`
		INSERT INTO agg_user (login, company, hide, type)
		VALUES ($1, $2, false, 'User')
	`, login, company)
	mustExec(`
		INSERT INTO agg_repo (owner, name, fork, stargazers_count, forks_count, language)
		VALUES ($1, 'repo', false, 5, 1, 'Go')
	`, login)

	if got := PopularDevs("User", ""); len(got) != 1 {
		t.Fatalf("expected 1 dev without company filter, got %d", len(got))
	}
	if got := PopularDevs("User", "acme"); len(got) != 1 {
		t.Fatalf("expected 1 dev with matching company filter, got %d", len(got))
	}
	if got := PopularDevs("User", "nonexistent"); len(got) != 0 {
		t.Fatalf("expected 0 devs with non-matching filter, got %d", len(got))
	}
}

func resetTables(t *testing.T) {
	t.Helper()
	if _, err := db.Exec("DELETE FROM agg_repo"); err != nil {
		t.Fatalf("failed to reset agg_repo: %v", err)
	}
	if _, err := db.Exec("DELETE FROM agg_user"); err != nil {
		t.Fatalf("failed to reset agg_user: %v", err)
	}
}

func mustExec(query string, args ...any) {
	if _, err := db.Exec(query, args...); err != nil {
		panic(err)
	}
}
