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
		where language is not null and fork=false
		group by language
		order by count desc
		limit 50;`

	queryPopularDevs = `
		select login, name, avatar_url, followers, public_repos, stars, forks, type
		from agg_user
		join (
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from agg_repo
			group by owner
		) repo ON (repo.owner=agg_user.login)
		where type='User'
		and hide is false
		order by stars desc
		limit 100;`

	queryPopularOrgs = `
		select login, name, avatar_url, followers, public_repos, stars, forks, type
		from agg_user
		join(
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from agg_repo
			group by owner
		) repo ON (repo.owner=agg_user.login)
		where type='Organization'
		and hide is false
		order by stars desc
		limit 100;`

	queryLanguage = `
		select * from (
			select owner, repo.user as user, repo.type as type, name, description, forks_count, stargazers_count, watchers_count, fork, (
				select sum(stargazers_count)
				from agg_repo
				where lower(language)=lower($1) and fork=false and owner=r1.owner
			) as count, row_number() over (partition by owner order by stargazers_count desc) as rownum
			from agg_repo r1
			join (
				select login, name as user, type
				from agg_user
			) repo ON (owner=login)
			where LOWER(r1.language)=LOWER($1) and r1.fork=false
			group by owner, name
			order by count desc, owner, stargazers_count desc
		) q where rownum < 4`

	queryProfileForUser = `
		select login, stars, forks, email, name, bio, blog, followers, public_repos, public_gists, avatar_url, hide, is_admin
		from agg_user
		join (
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from agg_repo
			group by owner
		) repo ON (repo.owner=agg_user.login)
		where login=$1`

	queryRepoForUser = `
		select name, fork, description, language, forks_count, stargazers_count
		from agg_repo
		where lower(owner)=lower($1) and language is not null
		order by language, stargazers_count desc, name`

	querySearchUsers = `
		select login, stars, forks, name, followers, public_repos, public_gists, avatar_url, type, hide, is_admin
		from agg_user
		join (
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from agg_repo
			group by owner
		) repo ON (repo.owner=agg_user.login)
		where hide is false and (
			login like LOWER($1) or 
			LOWER(name) like LOWER($1) or
			LOWER(bio) like LOWER($1) or
			LOWER(email) like LOWER($1)
		)
		order by stars desc
		limit 100`

	querySearchRepos = `
		select *
		from agg_repo
		where LOWER(name) like LOWER($1)
			or LOWER(description) like LOWER($1)
		order by stargazers_count desc
		limit 100
	`

	countLanguageUsers = `select count(distinct(owner))
			from agg_repo
			where lower(language)=lower($1) and fork=false;`
)
