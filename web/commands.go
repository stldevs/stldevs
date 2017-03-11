package web

import (
	"log"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
	"database/sql"
)

type Stldevs struct {
	*sqlx.DB
	*oauth2.Config
}

type AdminCommands interface {
}

const pageSize = 20

type Commands interface {
	// queries
	LastRun() *time.Time
	PopularLanguages() []LanguageCount
	PopularDevs() []DevCount
	PopularOrgs() []DevCount
	Language(name string, page int) ([]*LanguageResult, int)
	Profile(name string) (*ProfileData, error)
	Search(term, kind string) interface{}

	// auth flow
	AuthCode(string, oauth2.AuthCodeOption) string
	GithubLogin(code string) (*github.User, error)
}

func (s *Stldevs) LastRun() *time.Time {
	timeStr := mysql.NullTime{}
	err := s.Get(&timeStr, queryLastRun)
	if err == sql.ErrNoRows {
		return &time.Time{}
	}
	if err != nil {
		log.Println(err)
		return nil
	}
	if !timeStr.Valid {
		log.Println("null time in LastRun call results")
		return nil
	}
	return &timeStr.Time
}

type LanguageCount struct {
	Language string
	Count    int
	Users    int
}

func (s *Stldevs) PopularLanguages() []LanguageCount {
	langs := []LanguageCount{}
	err := s.Select(&langs, queryPopularLanguages)
	if err != nil {
		log.Println(err)
		return nil
	}
	return langs
}

type DevCount struct {
	Login, AvatarUrl, Followers, PublicRepos string
	Name                                     *string
	Stars, Forks                             int
}

func (s *Stldevs) PopularDevs() []DevCount {
	devs := []DevCount{}
	err := s.Select(&devs, queryPopularDevs)
	if err != nil {
		log.Println(err)
		return nil
	}
	return devs
}

func (s *Stldevs) PopularOrgs() []DevCount {
	devs := []DevCount{}
	err := s.Select(&devs, queryPopularOrgs)
	if err != nil {
		log.Println(err)
		return nil
	}
	return devs
}

type LanguageResult struct {
	Owner string
	Repos []stldevs.Repository
	Count int
}

func (s *Stldevs) Language(name string, page int) ([]*LanguageResult, int) {
	repos := []struct {
		stldevs.Repository
		Count int
	}{}
	offset := page * pageSize
	err := s.Select(&repos, queryLanguage, name, offset, pageSize, name)
	if err != nil {
		log.Println(err)
		return nil, 0
	}
	results := []*LanguageResult{}
	var cursor *LanguageResult
	for _, repo := range repos {
		if cursor == nil || cursor.Owner != *repo.Owner {
			cursor = &LanguageResult{Owner: *repo.Owner, Repos: []stldevs.Repository{repo.Repository}, Count: repo.Count}
			results = append(results, cursor)
		} else {
			cursor.Repos = append(cursor.Repos, repo.Repository)
		}
	}

	var total int
	if err = s.Get(&total, countLanguageUsers, name); err != nil {
		log.Println(err)
	}

	return results, total
}

type ProfileData struct {
	User  *github.User
	Repos map[string][]stldevs.Repository
}

func (s *Stldevs) Profile(name string) (*ProfileData, error) {
	user := &github.User{}
	reposByLang := map[string][]stldevs.Repository{}
	profile := &ProfileData{user, reposByLang}
	err := s.Get(profile.User, queryProfileForUser, name)
	if err != nil {
		log.Println("Error querying profile", name)
		return nil, err
	}

	if profile.User.Blog != nil && *profile.User.Blog != "" && !strings.HasPrefix(*profile.User.Blog, "http://") {
		*profile.User.Blog = "http://" + *profile.User.Blog
	}

	repos := []stldevs.Repository{}
	err = s.Select(&repos, queryRepoForUser, name)
	if err != nil {
		log.Println("Error querying repo for user", name)
		return nil, err
	}

	for _, repo := range repos {
		lang := *repo.Language
		if _, ok := reposByLang[lang]; !ok {
			reposByLang[lang] = []stldevs.Repository{repo}
			continue
		}
		reposByLang[lang] = append(reposByLang[lang], repo)
	}

	return profile, nil
}

func (s *Stldevs) Search(term, kind string) interface{} {
	query := "%" + term + "%"
	if kind == "users" {
		users := []stldevs.User{}
		if err := s.Select(&users, querySearchUsers, query, query); err != nil {
			log.Println(err)
			return nil
		}
		return users
	} else if kind == "repos" {
		repos := []stldevs.Repository{}
		if err := s.Select(&repos, querySearchRepos, query, query); err != nil {
			log.Println(err)
			return nil
		}
		return repos
	}
	log.Println("Unknown search kind", kind)
	return nil
}

func (s *Stldevs) AuthCode(state string, option oauth2.AuthCodeOption) string {
	return s.AuthCodeURL(state, option)
}

func (s *Stldevs) GithubLogin(code string) (*github.User, error) {
	token, err := s.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	client := github.NewClient(s.Client(oauth2.NoContext, token))

	user, _, err := client.Users.Get("")
	if err != nil {
		return nil, err
	}

	return user, nil
}
