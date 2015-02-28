package aggregator

import (
	"database/sql"
	"log"
	"os"

	"code.google.com/p/goauth2/oauth"

	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
)

type Aggregator struct {
	client  *github.Client
	db      *sql.DB
	running bool
}

func New(db *sql.DB) *Aggregator {
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
		Token: &oauth.Token{AccessToken: os.Getenv("GITHUB_API_KEY")},
	}

	var agg Aggregator
	agg.db = db
	agg.client = github.NewClient(t.Client())
	return &agg
}

func (a *Aggregator) Run() {
	a.running = true
	defer func() { a.running = false }()
	a.insertRunLog()
	step1(a)
	step2(a)
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
}

func (a *Aggregator) PopularLanguages() []LanguageCount {
	rows, err := a.db.Query(`
		select language, count(*) as count
		from agg_repo
		where language is not null
		group by language
		order by count desc;
	`)
	check(err)
	defer rows.Close()

	langs := []LanguageCount{}
	for rows.Next() {
		var lang string
		var count int
		if err = rows.Scan(&lang, &count); err != nil {
			log.Println(err)
		}
		langs = append(langs, LanguageCount{lang, count})
	}
	return langs
}

type LanguageResult struct {
	Owner *github.User
	Repos []github.Repository
	Count int
}

func (a *Aggregator) Language(name string) []*LanguageResult {
	rows, err := a.db.Query(`
SELECT r1.owner, r1.name, r1.description, r1.forks_count, r1.stargazers_count, r1.watchers_count, cnt
FROM agg_repo r1
JOIN (
	select owner, count(*) as cnt
	from stldevs.agg_repo
	where language=?
	group by owner
) r2 ON ( r2.owner = r1.owner )
where language=?
order by r2.cnt desc, stargazers_count desc;
`, name, name)
	check(err)
	defer rows.Close()

	data := []*LanguageResult{}
	var cur *LanguageResult
	for rows.Next() {
		repo := github.Repository{}
		repo.Owner = &github.User{}
		var count int
		rows.Scan(&repo.Owner.Login, &repo.Name, &repo.Description, &repo.ForksCount,
			&repo.StargazersCount, &repo.WatchersCount, &count)
		if cur == nil || *cur.Owner.Login != *repo.Owner.Login {
			cur = &LanguageResult{repo.Owner, []github.Repository{repo}, count}
			data = append(data, cur)
		} else {
			cur.Repos = append(cur.Repos, repo)
		}
	}

	return data
}
