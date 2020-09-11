package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
)

type migration func(*sqlx.DB) error

var migrations []migration

func init() {
	migrations = []migration{
		genesis,
		organizations,
		userEnhancements,
	}
}

func Migrate(db *sqlx.DB) error {
	for _, migration := range migrations {
		if err := migration(db); err != nil {
			return err
		}
	}
	return nil
}

type migrationRecord struct {
	Name string
}

func genesis(db *sqlx.DB) error {
	tables := []string{
		createMeta,
		createUser,
		createRepo,
		createMigrations,
	}

	for _, t := range tables {
		_, err := db.Exec(t)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func organizations(db *sqlx.DB) error {
	m := []migrationRecord{}
	if err := db.Select(&m, selectMigrations, "organizations"); err != nil {
		log.Println(err)
		return err
	}
	if len(m) == 1 {
		return nil
	}
	tx, err := db.Beginx()
	if err != nil {
		log.Println(err)
		return err
	}
	if _, err := tx.Exec(migrationOrganizations); err != nil {
		log.Println(err)
		return err
	}
	if _, err := tx.Exec(insertMigration, "organizations"); err != nil {
		log.Println(err)
		return err
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func userEnhancements(db *sqlx.DB) error {
	_, err := db.Exec("alter table agg_user add column if not exists hide boolean default false")
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = db.Exec("alter table agg_user add column if not exists is_admin boolean default false")
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = db.Exec(`ALTER TABLE agg_user ADD COLUMN IF NOT EXISTS refreshed_at TIMESTAMPTZ`)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = db.Exec(`ALTER TABLE agg_repo ADD COLUMN IF NOT EXISTS refreshed_at TIMESTAMPTZ`)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
