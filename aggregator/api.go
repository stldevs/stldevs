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
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(createUser)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(createRepo)
	if err != nil {
		log.Fatal(err)
	}

	t := &oauth.Transport{Token: &oauth.Token{AccessToken: githubKey}}
	return &Aggregator{db: db, client: github.NewClient(t.Client())}
}

func (a *Aggregator) Run() {
	if a.running {
		return
	}
	a.running = true
	defer func() { a.running = false }()
	if err := a.insertRunLog(); err != nil {
		log.Println(err)
		return
	}
	users, err := a.findStlUsers()
	if err != nil {
		log.Println(err)
		return
	}
	if err = a.updateUsers(users); err != nil {
		log.Println(err)
		return
	}
	if err = a.updateRepos(); err != nil {
		log.Println(err)
	}
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
	langs := []LanguageCount{}
	err := a.db.Select(&langs, queryPopularLanguages)
	if err != nil {
		log.Println(err)
		return nil
	}
	return langs
}

type DevCount struct {
	Login, Name, AvatarUrl, Followers string
	Stars                             int
	Forks                             int
}

func (a *Aggregator) PopularDevs() []DevCount {
	devs := []DevCount{}
	err := a.db.Select(&devs, queryPopularDevs)
	if err != nil {
		log.Println(err)
		return nil
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

func (a *Aggregator) Profile(name string) (*ProfileData, error) {
	rows, err := a.db.Query(queryProfileForUser, name)
	if err != nil {
		log.Println("Error querying profile")
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Println("No rows found for Profile", name)
		return nil, nil
	}

	profile := &ProfileData{&github.User{}, map[string][]github.Repository{}}
	user := profile.User
	err = rows.Scan(&user.Login, &user.Email, &user.Name, &user.Blog, &user.Followers, &user.PublicRepos,
		&user.PublicGists, &user.AvatarURL)
	if err != nil {
		log.Println("Error scanning profile data")
		return nil, err
	}

	if user.Blog != nil && *user.Blog != "" && !strings.HasPrefix(*user.Blog, "http://") {
		*user.Blog = "http://" + *user.Blog
	}

	rows, err = a.db.Query(queryRepoForUser, user.Login)
	if err != nil {
		log.Println("Error querying repo for user", user.Login)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var repo github.Repository
		err = rows.Scan(&repo.Name, &repo.Language, &repo.ForksCount, &repo.StargazersCount)
		if err != nil {
			log.Println("Error scanning repo row")
			return nil, err
		}

		if _, ok := profile.Repos[*repo.Language]; ok {
			profile.Repos[*repo.Language] = append(profile.Repos[*repo.Language], repo)
		} else {
			profile.Repos[*repo.Language] = []github.Repository{repo}
		}
	}

	return profile, nil
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
