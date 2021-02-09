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

var (
	// LastRun returns the last time github was scraped
	LastRun          = lastRun
	PopularLanguages = popularLanguages
	PopularDevs      = popularDevs
	PopularOrgs      = popularOrgs
	Language         = language
	Profile          = profile
	SearchUsers      = searchUsers
	SearchRepos      = searchRepos
	HideUser         = hideUser
	Delete           = deleteUser
	IsAdmin          = isAdmin
)

func lastRun() time.Time {
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

func popularLanguages() []LanguageCount {
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
	AvatarUrl   string  `json:"avatar_url"`
	Followers   string  `json:"followers"`
	PublicRepos string  `json:"public_repos"`
	Name        *string `json:"name"`
	Stars       int     `json:"stars"`
	Forks       int     `json:"forks"`
	Type        string  `json:"type"`
}

func popularDevs() []DevCount {
	devs := []DevCount{}
	err := db.Select(&devs, queryPopularDevs)
	if err != nil {
		log.Println(err)
		return nil
	}
	return devs
}

func popularOrgs() []DevCount {
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
	Name  *string `json:"name"`
	Type  string  `json:"type"`
}

var languageCache = struct {
	sync.RWMutex
	result  map[string][]*LanguageResult
	total   map[string]int
	lastRun time.Time
}{
	result: map[string][]*LanguageResult{},
	total:  map[string]int{},
}

func language(name string) ([]*LanguageResult, int) {
	lastRun := lastRun()
	languageCache.RLock()
	result, found := languageCache.result[name]
	if found && lastRun.Equal(languageCache.lastRun) {
		defer languageCache.RUnlock()
		return result, languageCache.total[name]
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
		return nil, 0
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

	var total int
	if err = db.Get(&total, countLanguageUsers, name); err != nil {
		log.Println(err)
	}

	languageCache.result[name] = results
	languageCache.total[name] = total
	languageCache.lastRun = lastRun
	return results, total
}

type StlDevsUser struct {
	*github.User
	Stars   int  `json:"stars"`
	Forks   int  `json:"forks"`
	Hide    bool `json:"hide,omitempty"`
	IsAdmin bool `json:"is_admin,omitempty"`
}

type ProfileData struct {
	User  *StlDevsUser
	Repos map[string][]stldevs.Repository
}

func profile(name string) (*ProfileData, error) {
	// TODO hide the user when other users try to see them but they are set to "Hide" in db

	// There are 2 queries to do so run them concurrently
	userCh := make(chan *StlDevsUser)
	reposCh := make(chan map[string][]stldevs.Repository)
	defer close(userCh)
	defer close(reposCh)

	go func() {
		user := &StlDevsUser{}
		err := db.Get(user, queryProfileForUser, name)
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
			lang := *repo.Language
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

func searchUsers(term string) []StlDevsUser {
	query := "%" + term + "%"
	users := []StlDevsUser{}
	if err := db.Select(&users, querySearchUsers, query); err != nil {
		log.Println(err)
		return nil
	}
	return users
}

func searchRepos(term string) []stldevs.Repository {
	query := "%" + term + "%"
	repos := []stldevs.Repository{}
	if err := db.Select(&repos, querySearchRepos, query); err != nil {
		log.Println(err)
		return nil
	}
	return repos
}

func hideUser(hide bool, login string) error {
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

func deleteUser(login string) error {
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

func isAdmin(login string) bool {
	rows, err := db.Query("select is_admin from agg_user where login=$1", login)
	if err != nil {
		log.Println(err)
	}

	var isAdmin bool
	if err == nil && rows.Next() && rows.Scan(&isAdmin) == nil && isAdmin == true {
		return true
	}
	return false
}
