package migrations

const (
	createMeta = `
		CREATE TABLE IF NOT EXISTS agg_meta (
			created_at TIMESTAMP);`

	createUser = `CREATE TABLE IF NOT EXISTS agg_user (
			login VARCHAR(255) NOT NULL PRIMARY KEY,
			email TEXT,
			location TEXT,
			hireable BOOL,
			blog TEXT,
			bio TEXT,
			followers INTEGER,
			following INTEGER,
			public_repos INTEGER,
			public_gists INTEGER,
			avatar_url TEXT,
			disk_usage INTEGER,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
			);`

	createRepo = `CREATE TABLE IF NOT EXISTS agg_repo (
			owner VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			language VARCHAR(255),
			homepage TEXT,
			forks_count INT,
			network_count INT,
			open_issues_count INT,
			stargazers_count INT,
			subscribers_count INT,
			watchers_count INT,
			size INT,
			fork BOOL,
			default_branch TEXT,
			master_branch TEXT,
			created_at TIMESTAMP,
			pushed_at TIMESTAMP,
			updated_at TIMESTAMP,
			primary key (owner, name)
			);`

	createMigrations = `CREATE TABLE IF NOT EXISTS migrations (
			name VARCHAR(255) NOT NULL PRIMARY KEY
			);`

	selectMigrations = "select * from migrations where name=$1"
	insertMigration  = `INSERT INTO migrations VALUES($1)`

	migrationOrganizations = `ALTER TABLE agg_user
		ADD COLUMN type VARCHAR(255),
		ADD COLUMN name VARCHAR(255)`
)
