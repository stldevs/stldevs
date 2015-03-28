package aggregator

import (
	"errors"
	"log"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
	"github.com/jmoiron/sqlx"
)

type Aggregator struct {
	client  *github.Client
	db      *sqlx.DB
	running bool
}

func New(db *sqlx.DB, githubKey string) *Aggregator {
	_, err := db.Exec(createMeta)
	check(err)
	_, err = db.Exec(createUser)
	check(err)
	_, err = db.Exec(createRepo)
	check(err)

	t := &oauth.Transport{Token: &oauth.Token{AccessToken: githubKey}}
	return &Aggregator{db: db, client: github.NewClient(t.Client())}
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
	a.updateRepos()
}

func (a *Aggregator) Running() bool {
	return a.running
}

func (a *Aggregator) LastRun() (string, error) {
	rows, err := a.db.Query(queryLastRun)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		// has never run!
		return time.Time{}.Local().Format("Jan 2, 2006 at 3:04pm"), nil
	}
	// it might be null
	var t mysql.NullTime
	if err = rows.Scan(&t); err != nil {
		log.Println(err)
		return "", err
	}
	if !t.Valid {
		err = errors.New("null time in LastRun call results")
		log.Println(err.Error())
		return "", err
	}
	return t.Time.Local().Format("Jan 2, 2006 at 3:04pm"), nil
}

type LanguageCount struct {
	Language string
	Count    int
	Users    int
}

func (a *Aggregator) PopularLanguages() []LanguageCount {
	rows, err := a.db.Query(queryPopularLanguages)
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
	rows, err := a.db.Query(queryPopularDevs)
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
	Owner string
	Repos []Repository
	Count int
}

func (a *Aggregator) Language(name string) []*LanguageResult {
	repos := []struct {
		Repository
		Count int
	}{}
	err := a.db.Select(&repos, queryLanguage, name, name)
	if err != nil {
		log.Println(err)
		return nil
	}
	results := []*LanguageResult{}
	var cursor *LanguageResult
	for _, repo := range repos {
		if cursor == nil || cursor.Owner != *repo.Owner {
			cursor = &LanguageResult{Owner: *repo.Owner, Repos: []Repository{repo.Repository}, Count: repo.Count}
			results = append(results, cursor)
		} else {
			cursor.Repos = append(cursor.Repos, repo.Repository)
		}
	}
	return results
}

type ProfileData struct {
	User  *github.User
	Repos map[string][]github.Repository
}

func (a *Aggregator) Profile(name string) *ProfileData {
	rows, err := a.db.Query(queryProfileForUser, name)
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
	rows.Close()

	if user.Blog != nil && *user.Blog != "" && !strings.HasPrefix(*user.Blog, "http://") {
		*user.Blog = "http://" + *user.Blog
	}

	rows, err = a.db.Query(queryRepoForUser, user.Login)
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

func (a *Aggregator) Search(term string) *[]User {
	query := "%" + term + "%"
	users := []User{}
	if err := a.db.Select(&users, querySearch, query, query); err != nil {
		log.Println(err)
		return nil
	}

	return &users
}
