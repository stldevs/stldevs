package migrations

const (
	createMeta = `
		CREATE TABLE IF NOT EXISTS agg_meta (
			created_at DATETIME);`

	createUser = `CREATE TABLE IF NOT EXISTS agg_user (
			login VARCHAR(255) NOT NULL PRIMARY KEY,
			email TEXT,
			name VARCHAR(255),
			location TEXT,
			hireable BOOL,
			blog TEXT,
			bio TEXT,
			followers INTEGER,
			following INTEGER,
			public_repos INTEGER,
			public_gists INTEGER,
			avatar_url TEXT,
			type VARCHAR(255),
			disk_usage INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			INDEX type (type),
			INDEX name (name)
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
			created_at DATETIME,
			pushed_at DATETIME,
			updated_at DATETIME,
			primary key (owner, name),
			INDEX language (language),
			INDEX forks_count (forks_count),
			INDEX network_count (network_count),
			INDEX stargazers_count (stargazers_count),
			INDEX subscribers_count (subscribers_count),
			INDEX watchers_count (watchers_count)
			);`

	createMigrations = `CREATE TABLE IF NOT EXISTS migrations (
			name VARCHAR(255) NOT NULL PRIMARY KEY
			);`

	selectMigrations = "select * from migrations where name=?"
	insertMigration  = `INSERT INTO migrations VALUES(?)`

	migrationOrganizations = `ALTER TABLE agg_user
		ADD COLUMN type VARCHAR(255) AFTER avatar_url,
		MODIFY name VARCHAR(255),
		ADD INDEX type (type),
		ADD INDEX name (name)`
)
