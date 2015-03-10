package aggregator

import (
	"database/sql"
	"log"

	"code.google.com/p/goauth2/oauth"

	"time"

	"strings"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
)

type Aggregator struct {
	client  *github.Client
	db      *sql.DB
	running bool
}

func New(db *sql.DB, githubKey string) *Aggregator {
	// TODO add more metadata like number of users found etc
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS agg_meta (
		created_at DATETIME
	);`)
	check(err)

	_, err = db.Exec(`
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
			updated_at DATETIME
		);`)
	check(err)
	_, err = db.Exec(`
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
			updated_at DATETIME
		);`)
	check(err)

	// init client
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: githubKey},
	}

	var agg Aggregator
	agg.db = db
	agg.client = github.NewClient(t.Client())
	return &agg
}

func (a *Aggregator) Run() {
	if a.running {
		return
	}
	a.running = true
	defer func() { a.running = false }()
	a.insertRunLog()
	users := a.findStlUsers()
	a.updateUsers(users)
	step2(a)
}

func (a *Aggregator) Running() bool {
	return a.running
}

func (a *Aggregator) LastRun() time.Time {
	rows, err := a.db.Query("select created_at from agg_meta order by created_at desc limit 1;")
	check(err)
	defer rows.Close()

	if !rows.Next() {
		return time.Time{}
	}
	var t mysql.NullTime
	if err = rows.Scan(&t); err != nil {
		log.Println(err)
	}
	return t.Time
}

type LanguageCount struct {
	Language string
	Count    int
	Users    int
}

func (a *Aggregator) PopularLanguages() []LanguageCount {
	rows, err := a.db.Query(`
		select language, count(*) as cnt, count(distinct(owner))
		from agg_repo
		where language is not null
		group by language
		order by cnt desc;
	`)
	check(err)
	defer rows.Close()

	langs := []LanguageCount{}
	for rows.Next() {
		var lang string
		var count int
		var owners int
		if err = rows.Scan(&lang, &count, &owners); err != nil {
			log.Println(err)
		} else {
			langs = append(langs, LanguageCount{lang, count, owners})
		}
	}
	return langs
}

type DevCount struct {
	Login, Name, AvatarUrl, Followers string
	Stars                             int
	Forks                             int
}

func (a *Aggregator) PopularDevs() []DevCount {
	rows, err := a.db.Query(`
		select login,name,avatar_url,followers,cnt,frks
		from stldevs.agg_user user
		join(
			select owner,sum(stargazers_count) as cnt,sum(forks_count) as frks
			from stldevs.agg_repo
			group by owner
		) repo ON (repo.owner=user.login)
		where name is not null and cnt > 100
		order by cnt desc;`)
	check(err)
	defer rows.Close()

	devs := []DevCount{}
	for rows.Next() {
		dev := DevCount{}
		if err = rows.Scan(&dev.Login, &dev.Name, &dev.AvatarUrl, &dev.Followers, &dev.Stars, &dev.Forks); err != nil {
			log.Println(err)
		} else {
			devs = append(devs, dev)
		}
	}
	return devs
}

type LanguageResult struct {
	Owner *github.User
	Repos []github.Repository
	Count int
}

func (a *Aggregator) Language(name string) []*LanguageResult {
	rows, err := a.db.Query(`
		SELECT r1.owner, r1.name, r1.description, r1.forks_count, r1.stargazers_count, r1.watchers_count, r1.fork, cnt
		FROM agg_repo r1
		JOIN (
			select owner, count(*) as cnt
			from stldevs.agg_repo
			where language=?
			group by owner
		) r2 ON ( r2.owner = r1.owner )
		where language=?
		order by r2.cnt desc, r2.owner, stargazers_count desc;`, name, name)
	check(err)
	defer rows.Close()

	data := []*LanguageResult{}
	var cur *LanguageResult
	for rows.Next() {
		repo := github.Repository{}
		repo.Owner = &github.User{}
		var count int
		rows.Scan(&repo.Owner.Login, &repo.Name, &repo.Description, &repo.ForksCount,
			&repo.StargazersCount, &repo.WatchersCount, &repo.Fork, &count)
		if cur == nil || *cur.Owner.Login != *repo.Owner.Login {
			cur = &LanguageResult{repo.Owner, []github.Repository{repo}, count}
			data = append(data, cur)
		} else {
			cur.Repos = append(cur.Repos, repo)
		}
	}

	return data
}

type ProfileData struct {
	User  *github.User
	Repos map[string][]github.Repository
}

func (a *Aggregator) Profile(name string) *ProfileData {
	rows, err := a.db.Query(`
	select login,email,name,blog,followers,public_repos,public_gists,avatar_url
	from agg_user
	where login=?`, name)
	check(err)
	defer rows.Close()

	if !rows.Next() {
		log.Println("No rows found for Profile", name)
		return nil
	}

	profile := &ProfileData{&github.User{}, map[string][]github.Repository{}}
	user := profile.User
	err = rows.Scan(&user.Login, &user.Email, &user.Name, &user.Blog, &user.Followers, &user.PublicRepos,
		&user.PublicGists, &user.AvatarURL)
	check(err)

	if user.Blog != nil && *user.Blog != "" && !strings.HasPrefix(*user.Blog, "http://") {
		*user.Blog = "http://" + *user.Blog
	}

	rows.Close()

	rows, err = a.db.Query(`
	select name,language,forks_count,stargazers_count
	from agg_repo
	where owner=? and language is not null
	order by language, stargazers_count desc, name`, user.Login)
	check(err)

	for rows.Next() {
		var repo github.Repository
		err = rows.Scan(&repo.Name, &repo.Language, &repo.ForksCount, &repo.StargazersCount)
		check(err)

		if _, ok := profile.Repos[*repo.Language]; ok {
			profile.Repos[*repo.Language] = append(profile.Repos[*repo.Language], repo)
		} else {
			profile.Repos[*repo.Language] = []github.Repository{repo}
		}
	}

	return profile
}

func (a *Aggregator) Search(query string) []github.User {
	percentified := "%" + query + "%"
	rows, err := a.db.Query(`
	select login,email,name,blog,followers,public_repos,public_gists,avatar_url
	from agg_user
	where login like ? or name like ?`, percentified, percentified)
	check(err)
	defer rows.Close()

	users := []github.User{}
	for rows.Next() {
		user := github.User{}
		err = rows.Scan(&user.Login, &user.Email, &user.Name, &user.Blog, &user.Followers, &user.PublicRepos,
			&user.PublicGists, &user.AvatarURL)
		check(err)
		users = append(users, user)
	}

	return users
}
