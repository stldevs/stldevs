package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs"
)

// LastRun returns the last time github was scraped
var LastRun = func() time.Time {
	var lastRun time.Time
	err := db.Get(&lastRun, queryLastRun)
	if err == sql.ErrNoRows {
		return lastRun
	}
	if err != nil {
		log.Println(err)
		return lastRun
	}
	return lastRun
}

type LanguageCount struct {
	Language string
	Count    int
	Users    int
}

var PopularLanguages = func() []LanguageCount {
	langs := []LanguageCount{}
	err := db.Select(&langs, queryPopularLanguages)
	if err != nil {
		log.Println(err)
		return nil
	}
	return langs
}

type DevCount struct {
	Login       string  `json:"login"`
	Company     string  `json:"company"`
	AvatarUrl   string  `json:"avatar_url"`
	Followers   string  `json:"followers"`
	PublicRepos string  `json:"public_repos"`
	Name        *string `json:"name"`
	Stars       int     `json:"stars"`
	Forks       int     `json:"forks"`
	Type        string  `json:"type"`
}

var PopularDevs = func(devType, company string) []DevCount {
	devs := []DevCount{}
	arguments := []interface{}{devType}

	query := `select login, name, company, avatar_url, followers, public_repos, stars, forks, type
		from agg_user
		join (
			select owner, sum(stargazers_count) as stars, sum(forks_count) as forks
			from agg_repo
			group by owner
		) repo ON (repo.owner=agg_user.login)
		where type=$1
		and hide is false
		`

	if company != "" {
		query += "and LOWER(company) like LOWER($2)\n"
		arguments = append(arguments, "%"+company+"%")
	}

	query += `order by stars desc limit 100`

	err := db.Select(&devs, query, arguments...)
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
	Name  *string `json:"name"`
	Type  string  `json:"type"`
}

var languageCache = struct {
	sync.RWMutex
	result  map[string][]*LanguageResult
	lastRun time.Time
}{
	result: map[string][]*LanguageResult{},
}

var Language = func(name string) []*LanguageResult {
	run := LastRun()
	languageCache.RLock()
	result, found := languageCache.result[name]
	if found && run.Equal(languageCache.lastRun) {
		defer languageCache.RUnlock()
		return result
	}
	languageCache.RUnlock()
	languageCache.Lock()
	defer languageCache.Unlock()

	var repos []struct {
		stldevs.Repository
		Count  int
		Rownum int
		Login  string  // not used, just here to satisfy sqlx
		User   *string `json:"user"`
		Type   string  `json:"type"`
	}
	err := db.Select(&repos, queryLanguage, name)
	if err != nil {
		log.Println(err)
		return nil
	}
	results := []*LanguageResult{}
	var cursor *LanguageResult
	for _, repo := range repos {
		if cursor == nil || cursor.Owner != *repo.Owner {
			cursor = &LanguageResult{
				Owner: *repo.Owner,
				Repos: []stldevs.Repository{repo.Repository},
				Count: repo.Count,
				Name:  repo.User,
				Type:  repo.Type,
			}
			results = append(results, cursor)
		} else {
			cursor.Repos = append(cursor.Repos, repo.Repository)
		}
	}

	languageCache.result[name] = results
	languageCache.lastRun = run
	return results
}

type StlDevsUser struct {
	*github.User
	Stars   int  `json:"stars"`
	Forks   int  `json:"forks"`
	Hide    bool `json:"hide,omitempty"`
	IsAdmin bool `json:"is_admin,omitempty"`
}

func GetUser(login string) (*StlDevsUser, error) {
	user := &StlDevsUser{}
	err := db.Get(user, queryUser, login)
	if err != nil {
		log.Println("Error querying user", login, err)
		return nil, err
	}
	return user, err
}

type ProfileData struct {
	User  *StlDevsUser
	Repos map[string][]stldevs.Repository
}

var Profile = func(name string) (*ProfileData, error) {
	// TODO hide the user when other users try to see them but they are set to "Hide" in db

	// There are 2 queries to do so run them concurrently
	userCh := make(chan *StlDevsUser)
	reposCh := make(chan map[string][]stldevs.Repository)
	defer close(userCh)
	defer close(reposCh)

	go func() {
		user := &StlDevsUser{}
		err := db.Get(user, queryUser, name)
		if err != nil {
			log.Println("Error querying profile", name, err)
			userCh <- nil
			return
		}
		userCh <- user
	}()

	go func() {
		repos := []stldevs.Repository{}
		err := db.Select(&repos, queryRepoForUser, name)
		if err != nil {
			log.Println("Error querying repo for user", name)
			reposCh <- nil
			return
		}

		reposByLang := map[string][]stldevs.Repository{}
		for _, repo := range repos {
			var lang string
			if repo.Language != nil {
				lang = *repo.Language
			}
			if _, ok := reposByLang[lang]; !ok {
				reposByLang[lang] = []stldevs.Repository{repo}
				continue
			}
			reposByLang[lang] = append(reposByLang[lang], repo)
		}

		reposCh <- reposByLang
	}()

	user := <-userCh
	repoMap := <-reposCh

	if user == nil || repoMap == nil {
		return nil, fmt.Errorf("not found")
	}

	for _, repos := range repoMap {
		for _, repo := range repos {
			user.Stars += *repo.StargazersCount
			user.Forks += *repo.ForksCount
		}
	}

	return &ProfileData{user, repoMap}, nil
}

var SearchUsers = func(term string) []StlDevsUser {
	query := "%" + term + "%"
	users := []StlDevsUser{}
	if err := db.Select(&users, querySearchUsers, query); err != nil {
		log.Println(err)
		return nil
	}
	return users
}

var SearchRepos = func(term string) []stldevs.Repository {
	query := "%" + term + "%"
	repos := []stldevs.Repository{}
	if err := db.Select(&repos, querySearchRepos, query); err != nil {
		log.Println(err)
		return nil
	}
	return repos
}

var HideUser = func(hide bool, login string) error {
	result, err := db.Exec("update agg_user set hide=$1 where login=$2", hide, login)
	if err != nil {
		log.Println(err)
		return err
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		return fmt.Errorf("affected no users")
	}
	return nil
}

var Delete = func(login string) error {
	_, err := db.Exec("delete from agg_repo where owner=$1", login)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = db.Exec("delete from agg_user where login=$1", login)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}
