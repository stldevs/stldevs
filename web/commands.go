package web

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

func LastRun(db *sqlx.DB) *time.Time {
	var lastRun time.Time
	err := db.Get(&lastRun, queryLastRun)
	if err == sql.ErrNoRows {
		return &time.Time{}
	}
	if err != nil {
		log.Println(err)
		return nil
	}
	if lastRun.Equal(time.Time{}) {
		log.Println("null time in LastRun call results")
		return nil
	}
	return &lastRun
}

type LanguageCount struct {
	Language string
	Count    int
	Users    int
}

func PopularLanguages(db *sqlx.DB) []LanguageCount {
	langs := []LanguageCount{}
	err := db.Select(&langs, queryPopularLanguages)
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

func PopularDevs(db *sqlx.DB) []DevCount {
	devs := []DevCount{}
	err := db.Select(&devs, queryPopularDevs)
	if err != nil {
		log.Println(err)
		return nil
	}
	return devs
}

func PopularOrgs(db *sqlx.DB) []DevCount {
	devs := []DevCount{}
	err := db.Select(&devs, queryPopularOrgs)
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

func Language(db *sqlx.DB, name string) ([]*LanguageResult, int) {
	repos := []struct {
		stldevs.Repository
		Count int
		Rownum int
	}{}
	err := db.Select(&repos, queryLanguage, name)
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
	if err = db.Get(&total, countLanguageUsers, name); err != nil {
		log.Println(err)
	}

	return results, total
}

type ProfileData struct {
	User  *github.User
	Repos map[string][]stldevs.Repository
}

func Profile(db *sqlx.DB, name string) (*ProfileData, error) {
	user := &github.User{}
	reposByLang := map[string][]stldevs.Repository{}
	profile := &ProfileData{user, reposByLang}
	err := db.Get(profile.User, queryProfileForUser, name)
	if err != nil {
		log.Println("Error querying profile", name)
		return nil, err
	}

	repos := []stldevs.Repository{}
	err = db.Select(&repos, queryRepoForUser, name)
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

func Search(db *sqlx.DB, term, kind string) interface{} {
	query := "%" + term + "%"
	if kind == "users" {
		users := []stldevs.User{}
		if err := db.Select(&users, querySearchUsers, query); err != nil {
			log.Println(err)
			return nil
		}
		return users
	} else if kind == "repos" {
		repos := []stldevs.Repository{}
		if err := db.Select(&repos, querySearchRepos, query); err != nil {
			log.Println(err)
			return nil
		}
		return repos
	}
	log.Println("Unknown search kind", kind)
	return nil
}

func AuthCode(conf *oauth2.Config, state string, option oauth2.AuthCodeOption) string {
	return conf.AuthCodeURL(state, option)
}

func GithubLogin(conf *oauth2.Config, code string) (*github.User, error) {
	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	client := github.NewClient(conf.Client(oauth2.NoContext, token))

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return nil, err
	}

	return user, nil
}
