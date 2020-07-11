package web

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/jakecoffman/stldevs"
	"github.com/jmoiron/sqlx"
)

func LastRun(db *sqlx.DB) time.Time {
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

var LanguageCache = struct {
	sync.RWMutex
	result  map[string][]*LanguageResult
	total   map[string]int
	lastRun time.Time
}{
	result: map[string][]*LanguageResult{},
	total:  map[string]int{},
}

func Language(db *sqlx.DB, name string) ([]*LanguageResult, int) {
	lastRun := LastRun(db)
	LanguageCache.RLock()
	result, found := LanguageCache.result[name]
	if found && lastRun.Equal(LanguageCache.lastRun) {
		defer LanguageCache.RUnlock()
		return result, LanguageCache.total[name]
	}
	LanguageCache.RUnlock()
	LanguageCache.Lock()
	defer LanguageCache.Unlock()

	repos := []struct {
		stldevs.Repository
		Count  int
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

	LanguageCache.result[name] = results
	LanguageCache.total[name] = total
	LanguageCache.lastRun = lastRun
	return results, total
}

type StlDevsUser struct {
	*github.User
	Stars   int  `json:",omitempty"`
	Forks   int  `json:",omitempty"`
	Hide    bool `json:",omitempty"`
	IsAdmin bool `json:",omitempty"`
}

type ProfileData struct {
	User  *StlDevsUser
	Repos map[string][]stldevs.Repository
}

func Profile(db *sqlx.DB, name string) (*ProfileData, error) {
	// TODO hide the user when other users try to see them but they are set to "Hide" in db

	// There are 2 queries to do so run them concurrently
	userCh := make(chan *StlDevsUser)
	reposCh := make(chan map[string][]stldevs.Repository)
	defer close(userCh)
	defer close(reposCh)

	go func() {
		start := time.Now()
		user := &StlDevsUser{}
		err := db.Get(user, queryProfileForUser, name)
		if err != nil {
			log.Println("Error querying profile", name)
			userCh <- nil
			return
		}
		log.Println("queryProfileForUser took", time.Now().Sub(start))
		userCh <- user
	}()

	go func() {
		start := time.Now()
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

		log.Println("queryRepoForUser took", time.Now().Sub(start))
		reposCh <- reposByLang
	}()

	user := <-userCh
	repos := <-reposCh

	if user == nil || repos == nil {
		return nil, fmt.Errorf("not found")
	}

	return &ProfileData{user, repos}, nil
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
