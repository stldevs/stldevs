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
		from agg_user
		join (
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from agg_repo
			group by owner
		) repo ON (repo.owner=agg_user.login)
		where type='User'
		order by stars desc
		limit 100;`

	queryPopularOrgs = `
		select login, name, avatar_url, followers, public_repos, stars, forks
		from agg_user
		join(
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from agg_repo
			group by owner
		) repo ON (repo.owner=agg_user.login)
		where type='Organization'
		order by stars desc
		limit 100;`

	queryLanguage = `
		SELECT r1.owner, r1.name, r1.description, r1.forks_count, r1.stargazers_count, r1.watchers_count, r1.fork, count
		FROM agg_repo r1
		JOIN (
			select owner, sum(stargazers_count) as count
			from agg_repo
			where language=$1 and fork=false
			group by owner
			order by count desc
			limit $2
			offset $3
		) r2 ON ( r2.owner = r1.owner )
		where language=$1 and fork=false
		order by r2.count desc, r2.owner, stargazers_count desc`

	queryProfileForUser = `
		select login, email, name, blog, followers, public_repos, public_gists, avatar_url
		from agg_user
		where login=$1`

	queryRepoForUser = `
		select name, fork, description, language, forks_count, stargazers_count
		from agg_repo
		where owner=$1 and language is not null
		order by language, stargazers_count desc, name`

	querySearchUsers = `
		select *
		from agg_user
		where login like LOWER($1)
			or LOWER(name) like LOWER($1)
			limit 100`

	querySearchRepos = `
		select *
		from agg_repo
		where LOWER(name) like LOWER($1)
			or LOWER(description) like LOWER($1)
			order by stargazers_count desc
			limit 100
	`

	countLanguageUsers = `select count(distinct owner)
			from agg_repo
			where language=$1 and fork=0;`
)
