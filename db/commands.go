package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jakecoffman/stldevs/db/sqlc"
)

// LastRun returns the last time github was scraped.
var LastRun = func() time.Time {
	if queries == nil {
		return time.Time{}
	}
	lastRun, err := queries.LastRun(context.Background())
	if err == sql.ErrNoRows {
		return time.Time{}
	}
	if err != nil {
		log.Println("LastRun query failed:", err)
		return time.Time{}
	}
	return lastRun
}

var PopularLanguages = func() []sqlc.PopularLanguagesRow {
	if queries == nil {
		return nil
	}
	rows, err := queries.PopularLanguages(context.Background())
	if err != nil {
		log.Println("PopularLanguages query failed:", err)
		return nil
	}
	return rows
}

var PopularDevs = func(devType, company string) []sqlc.PopularDevsRow {
	if queries == nil {
		return nil
	}
	params := sqlc.PopularDevsParams{
		DevType: sql.NullString{String: devType, Valid: devType != ""},
	}
	if company != "" {
		params.CompanyPattern = sql.NullString{String: "%" + company + "%", Valid: true}
	} else {
		params.CompanyPattern = sql.NullString{Valid: false}
	}
	rows, err := queries.PopularDevs(context.Background(), params)
	if err != nil {
		log.Println("PopularDevs query failed:", err)
		return nil
	}
	return rows
}

type LanguageResult struct {
	Owner string
	Repos []sqlc.LanguageLeadersRow
	Count int
	Name  string `json:"name"`
	Type  string `json:"type"`
}

var languageCache = struct {
	sync.RWMutex
	result  map[string][]*LanguageResult
	lastRun time.Time
}{
	result: map[string][]*LanguageResult{},
}

var Language = func(name string) []*LanguageResult {
	if queries == nil {
		return nil
	}
	run := LastRun()
	languageCache.RLock()
	result, found := languageCache.result[name]
	if found && run.Equal(languageCache.lastRun) {
		languageCache.RUnlock()
		return result
	}
	languageCache.RUnlock()
	languageCache.Lock()
	defer languageCache.Unlock()

	rows, err := queries.LanguageLeaders(context.Background(), name)
	if err != nil {
		log.Println("LanguageLeaders query failed:", err)
		return nil
	}
	var cursor *LanguageResult
	results := make([]*LanguageResult, 0, len(rows))
	for _, row := range rows {
		if cursor == nil || cursor.Owner != row.Owner {
			cursor = &LanguageResult{
				Owner: row.Owner,
				Repos: []sqlc.LanguageLeadersRow{row},
				Count: int(row.TotalStars),
				Name:  row.DisplayName,
				Type:  row.Type,
			}
			results = append(results, cursor)
			continue
		}
		cursor.Repos = append(cursor.Repos, row)
	}
	languageCache.result[name] = results
	languageCache.lastRun = run
	return results
}

func GetUser(login string) (sqlc.GetUserRow, error) {
	if queries == nil {
		return sqlc.GetUserRow{}, fmt.Errorf("database not initialized")
	}
	row, err := queries.GetUser(context.Background(), login)
	if err != nil {
		log.Println("Error querying user", login, err)
		return sqlc.GetUserRow{}, err
	}
	return row, nil
}

type ProfileData struct {
	User  sqlc.GetUserRow                   `json:"user"`
	Repos map[string][]sqlc.ReposForUserRow `json:"repos"`
}

var Profile = func(name string) (*ProfileData, error) {
	if queries == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	userCh := make(chan sqlc.GetUserRow)
	reposCh := make(chan map[string][]sqlc.ReposForUserRow)
	defer close(userCh)
	defer close(reposCh)

	go func() {
		user, err := GetUser(name)
		if err != nil {
			userCh <- sqlc.GetUserRow{}
			return
		}
		userCh <- user
	}()

	go func() {
		rows, err := queries.ReposForUser(context.Background(), name)
		if err != nil {
			log.Println("Error querying repo for user", name, err)
			reposCh <- nil
			return
		}
		repos := map[string][]sqlc.ReposForUserRow{}
		for _, repo := range rows {
			lang := repo.Language
			repos[lang] = append(repos[lang], repo)
		}
		reposCh <- repos
	}()

	user := <-userCh
	repoMap := <-reposCh
	if user.Login == "" || repoMap == nil {
		return nil, fmt.Errorf("not found")
	}

	return &ProfileData{User: user, Repos: repoMap}, nil
}

var SearchUsers = func(term string) []sqlc.SearchUsersRow {
	if queries == nil {
		return nil
	}
	pattern := "%" + term + "%"
	rows, err := queries.SearchUsers(context.Background(), pattern)
	if err != nil {
		log.Println("SearchUsers query failed:", err)
		return nil
	}
	return rows
}

var SearchRepos = func(term string) []sqlc.SearchReposRow {
	if queries == nil {
		return nil
	}
	pattern := "%" + term + "%"
	rows, err := queries.SearchRepos(context.Background(), pattern)
	if err != nil {
		log.Println("SearchRepos query failed:", err)
		return nil
	}
	return rows
}

var HideUser = func(hide bool, login string) error {
	if queries == nil {
		return fmt.Errorf("database not initialized")
	}
	affected, err := queries.HideUser(context.Background(), sqlc.HideUserParams{Hide: hide, Login: login})
	if err != nil {
		log.Println("HideUser update failed:", err)
		return err
	}
	if affected != 1 {
		return fmt.Errorf("affected no users")
	}
	return nil
}

var Delete = func(login string) error {
	if queries == nil {
		return fmt.Errorf("database not initialized")
	}
	if err := queries.DeleteReposByOwner(context.Background(), login); err != nil {
		log.Println("Failed deleting repos for", login, err)
		return err
	}
	if err := queries.DeleteUser(context.Background(), login); err != nil {
		log.Println("Failed deleting user", login, err)
		return err
	}
	return nil
}
