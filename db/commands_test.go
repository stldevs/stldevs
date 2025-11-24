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
	result := PopularDevs("User", "company", "")
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

	if got := PopularDevs("User", "", ""); len(got) != 1 {
		t.Fatalf("expected 1 dev without company filter, got %d", len(got))
	}
	if got := PopularDevs("User", "acme", ""); len(got) != 1 {
		t.Fatalf("expected 1 dev with matching company filter, got %d", len(got))
	}
	if got := PopularDevs("User", "nonexistent", ""); len(got) != 0 {
		t.Fatalf("expected 0 devs with non-matching filter, got %d", len(got))
	}
}

func TestPopularDevsSorting(t *testing.T) {
	resetTables(t)

	// Insert users with different stats
	// user1: high stars (100), low forks (10), medium followers (50), medium repos (30)
	mustExec(`
		INSERT INTO agg_user (login, company, hide, type, followers, public_repos)
		VALUES ('user1', '', false, 'User', 50, 30)
	`)
	mustExec(`
		INSERT INTO agg_repo (owner, name, fork, stargazers_count, forks_count, language)
		VALUES ('user1', 'repo1', false, 100, 10, 'Go')
	`)

	// user2: medium stars (50), high forks (80), high followers (100), low repos (10)
	mustExec(`
		INSERT INTO agg_user (login, company, hide, type, followers, public_repos)
		VALUES ('user2', '', false, 'User', 100, 10)
	`)
	mustExec(`
		INSERT INTO agg_repo (owner, name, fork, stargazers_count, forks_count, language)
		VALUES ('user2', 'repo2', false, 50, 80, 'Go')
	`)

	// user3: low stars (20), medium forks (30), low followers (20), high repos (100)
	mustExec(`
		INSERT INTO agg_user (login, company, hide, type, followers, public_repos)
		VALUES ('user3', '', false, 'User', 20, 100)
	`)
	mustExec(`
		INSERT INTO agg_repo (owner, name, fork, stargazers_count, forks_count, language)
		VALUES ('user3', 'repo3', false, 20, 30, 'Go')
	`)

	t.Run("SortByStars", func(t *testing.T) {
		// Default sorting by stars (descending)
		got := PopularDevs("User", "", "stars")
		if len(got) != 3 {
			t.Fatalf("expected 3 devs, got %d", len(got))
		}
		// Expected order: user1 (100), user2 (50), user3 (20)
		if got[0].Login != "user1" {
			t.Errorf("expected user1 first, got %s", got[0].Login)
		}
		if got[1].Login != "user2" {
			t.Errorf("expected user2 second, got %s", got[1].Login)
		}
		if got[2].Login != "user3" {
			t.Errorf("expected user3 third, got %s", got[2].Login)
		}
	})

	t.Run("SortByStarsDefault", func(t *testing.T) {
		// Empty string defaults to stars
		got := PopularDevs("User", "", "")
		if len(got) != 3 {
			t.Fatalf("expected 3 devs, got %d", len(got))
		}
		// Expected order: user1 (100), user2 (50), user3 (20)
		if got[0].Login != "user1" {
			t.Errorf("expected user1 first, got %s", got[0].Login)
		}
		if got[1].Login != "user2" {
			t.Errorf("expected user2 second, got %s", got[1].Login)
		}
		if got[2].Login != "user3" {
			t.Errorf("expected user3 third, got %s", got[2].Login)
		}
	})

	t.Run("SortByForks", func(t *testing.T) {
		got := PopularDevs("User", "", "forks")
		if len(got) != 3 {
			t.Fatalf("expected 3 devs, got %d", len(got))
		}
		// Expected order: user2 (80), user3 (30), user1 (10)
		if got[0].Login != "user2" {
			t.Errorf("expected user2 first, got %s", got[0].Login)
		}
		if got[1].Login != "user3" {
			t.Errorf("expected user3 second, got %s", got[1].Login)
		}
		if got[2].Login != "user1" {
			t.Errorf("expected user1 third, got %s", got[2].Login)
		}
	})

	t.Run("SortByFollowers", func(t *testing.T) {
		got := PopularDevs("User", "", "followers")
		if len(got) != 3 {
			t.Fatalf("expected 3 devs, got %d", len(got))
		}
		// Expected order: user2 (100), user1 (50), user3 (20)
		if got[0].Login != "user2" {
			t.Errorf("expected user2 first, got %s", got[0].Login)
		}
		if got[1].Login != "user1" {
			t.Errorf("expected user1 second, got %s", got[1].Login)
		}
		if got[2].Login != "user3" {
			t.Errorf("expected user3 third, got %s", got[2].Login)
		}
	})

	t.Run("SortByPublicRepos", func(t *testing.T) {
		got := PopularDevs("User", "", "public_repos")
		if len(got) != 3 {
			t.Fatalf("expected 3 devs, got %d", len(got))
		}
		// Expected order: user3 (100), user1 (30), user2 (10)
		if got[0].Login != "user3" {
			t.Errorf("expected user3 first, got %s", got[0].Login)
		}
		if got[1].Login != "user1" {
			t.Errorf("expected user1 second, got %s", got[1].Login)
		}
		if got[2].Login != "user2" {
			t.Errorf("expected user2 third, got %s", got[2].Login)
		}
	})

	t.Run("InvalidSortDefaultsToStars", func(t *testing.T) {
		got := PopularDevs("User", "", "invalid_sort")
		if len(got) != 3 {
			t.Fatalf("expected 3 devs, got %d", len(got))
		}
		// Should default to stars: user1 (100), user2 (50), user3 (20)
		if got[0].Login != "user1" {
			t.Errorf("expected user1 first with invalid sort, got %s", got[0].Login)
		}
	})
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
