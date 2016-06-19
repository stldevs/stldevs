package web

const (
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
		order by count desc
		limit 100;`

	queryPopularDevs = `
		select login, name, avatar_url, followers, public_repos, stars, forks
		from stldevs.agg_user user
		join(
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from stldevs.agg_repo
			group by owner
		) repo ON (repo.owner=user.login)
		where type='User'
		order by stars desc
		limit 100;`

	queryPopularOrgs = `
		select login, name, avatar_url, followers, public_repos, stars, forks
		from stldevs.agg_user user
		join(
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from stldevs.agg_repo
			group by owner
		) repo ON (repo.owner=user.login)
		where type='Organization'
		order by stars desc
		limit 100;`

	queryLanguage = `
		SELECT r1.owner, r1.name, r1.description, r1.forks_count, r1.stargazers_count, r1.watchers_count, r1.fork, count
		FROM agg_repo r1
		JOIN (
			select owner, sum(stargazers_count) as count
			from stldevs.agg_repo
			where language=? and fork=0
			group by owner
			order by count desc
			limit ?, ?
		) r2 ON ( r2.owner = r1.owner )
		where language=? and fork=0
		order by r2.count desc, r2.owner, stargazers_count desc`

	queryProfileForUser = `
		select login, email, name, blog, followers, public_repos, public_gists, avatar_url
		from agg_user
		where login=?`

	queryRepoForUser = `
		select name, fork, description, language, forks_count, stargazers_count
		from agg_repo
		where owner=? and language is not null
		order by language, stargazers_count desc, name`

	querySearchUsers = `
		select *
		from agg_user
		where login like LOWER(?)
			or LOWER(name) like LOWER(?)
			limit 100`

	querySearchRepos = `
		select *
		from agg_repo
		where LOWER(name) like LOWER(?)
			or LOWER(description) like LOWER(?)
			order by stargazers_count desc
			limit 100
	`

	countLanguageUsers = `select count(distinct owner)
			from stldevs.agg_repo
			where language=? and fork=0;`
)
