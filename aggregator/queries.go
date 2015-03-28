package aggregator

const (
	createMeta = `
		CREATE TABLE IF NOT EXISTS agg_meta (
			created_at DATETIME);`

	createUser = `
		CREATE TABLE IF NOT EXISTS agg_user (
			login VARCHAR(255) NOT NULL PRIMARY KEY,
			email TEXT,
			name TEXT,
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
			created_at DATETIME,
			updated_at DATETIME);`

	createRepo = `
		CREATE TABLE IF NOT EXISTS agg_repo (
			owner VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL PRIMARY KEY,
			description TEXT,
			language TEXT,
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
			updated_at DATETIME);`

	queryLastRun = `
		select created_at
		from agg_meta
		order by created_at desc
		limit 1;`

	queryPopularLanguages = `
		select language, count(*) as count, count(distinct(owner)) as users
		from agg_repo
		where language is not null
		group by language
		order by count desc;`

	queryPopularDevs = `
		select login, name, avatar_url, followers, stars, forks
		from stldevs.agg_user user
		join(
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from stldevs.agg_repo
			group by owner
		) repo ON (repo.owner=user.login)
		where name is not null and stars > 100
		order by stars desc;`

	queryLanguage = `
		SELECT r1.owner, r1.name, r1.description, r1.forks_count, r1.stargazers_count, r1.watchers_count, r1.fork, count
		FROM agg_repo r1
		JOIN (
			select owner, sum(stargazers_count) as count
			from stldevs.agg_repo
			where language=?
			group by owner
		) r2 ON ( r2.owner = r1.owner )
		where language=?
		order by r2.count desc, r2.owner, stargazers_count desc;`

	queryProfileForUser = `
		select login, email, name, blog, followers, public_repos, public_gists, avatar_url
		from agg_user
		where login=?`

	queryRepoForUser = `
		select name, language, forks_count, stargazers_count
		from agg_repo
		where owner=? and language is not null
		order by language, stargazers_count desc, name`

	querySearch = `
		select *
		from agg_user
		where login like ? or name like ?`
)
