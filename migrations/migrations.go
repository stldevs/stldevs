package migrations

import (
	"database/sql"
	"log"
)

type migration func(*sql.DB) error

var migrations []migration

func init() {
	migrations = []migration{
		genesis,
		organizations,
		userEnhancements,
	}
}

func Migrate(db *sql.DB) error {
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

func genesis(db *sql.DB) error {
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

func organizations(db *sql.DB) error {
	const name = "organizations"
	apply, err := shouldApply(db, name)
	if err != nil {
		log.Println(err)
		return err
	}
	if !apply {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(migrationOrganizations); err != nil {
		log.Println(err)
		return err
	}
	if _, err := tx.Exec(insertMigration, name); err != nil {
		log.Println(err)
		return err
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func userEnhancements(db *sql.DB) error {
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
	_, err = db.Exec(`ALTER TABLE agg_user ADD COLUMN IF NOT EXISTS company text not null`)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = db.Exec(`ALTER TABLE agg_user ADD COLUMN IF NOT EXISTS company text not null`)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func shouldApply(db *sql.DB, name string) (bool, error) {
	var existing migrationRecord
	err := db.QueryRow(selectMigrations, name).Scan(&existing.Name)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}
