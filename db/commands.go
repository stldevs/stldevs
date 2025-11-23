package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/go-github/v52/github"
	"github.com/jakecoffman/stldevs"
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

type LanguageCount struct {
	Language string
	Count    int
	Users    int
}

var PopularLanguages = func() []LanguageCount {
	if queries == nil {
		return nil
	}
	rows, err := queries.PopularLanguages(context.Background())
	if err != nil {
		log.Println("PopularLanguages query failed:", err)
		return nil
	}
	counts := make([]LanguageCount, 0, len(rows))
	for _, row := range rows {
		if !row.Language.Valid {
			continue
		}
		counts = append(counts, LanguageCount{
			Language: row.Language.String,
			Count:    int(row.RepoCount),
			Users:    int(row.UserCount),
		})
	}
	return counts
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
	devs := make([]DevCount, 0, len(rows))
	for _, row := range rows {
		devs = append(devs, DevCount{
			Login:       row.Login,
			Company:     row.Company,
			AvatarUrl:   nullStringValue(row.AvatarUrl),
			Followers:   formatInt(row.Followers),
			PublicRepos: formatInt(row.PublicRepos),
			Name:        nullStringPtr(row.Name),
			Stars:       int(row.Stars),
			Forks:       int(row.Forks),
			Type:        row.Type.String,
		})
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
		repo := repoFromLanguageRow(row)
		if cursor == nil || cursor.Owner != row.Owner {
			cursor = &LanguageResult{
				Owner: row.Owner,
				Repos: []stldevs.Repository{repo},
				Count: int(row.TotalStars),
				Name:  nullStringPtr(row.DisplayName),
				Type:  row.Type.String,
			}
			results = append(results, cursor)
			continue
		}
		cursor.Repos = append(cursor.Repos, repo)
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
	if queries == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	row, err := queries.GetUser(context.Background(), login)
	if err != nil {
		log.Println("Error querying user", login, err)
		return nil, err
	}
	return userFromRow(row), nil
}

type ProfileData struct {
	User  *StlDevsUser
	Repos map[string][]stldevs.Repository
}

var Profile = func(name string) (*ProfileData, error) {
	if queries == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	userCh := make(chan *StlDevsUser)
	reposCh := make(chan map[string][]stldevs.Repository)
	defer close(userCh)
	defer close(reposCh)

	go func() {
		user, err := GetUser(name)
		if err != nil {
			userCh <- nil
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
		repos := map[string][]stldevs.Repository{}
		for _, repo := range convertAggRepos(rows) {
			lang := ""
			if repo.Language != nil {
				lang = *repo.Language
			}
			repos[lang] = append(repos[lang], repo)
		}
		reposCh <- repos
	}()

	user := <-userCh
	repoMap := <-reposCh
	if user == nil || repoMap == nil {
		return nil, fmt.Errorf("not found")
	}

	for _, repos := range repoMap {
		for _, repo := range repos {
			if repo.StargazersCount != nil {
				user.Stars += *repo.StargazersCount
			}
			if repo.ForksCount != nil {
				user.Forks += *repo.ForksCount
			}
		}
	}

	return &ProfileData{User: user, Repos: repoMap}, nil
}

var SearchUsers = func(term string) []StlDevsUser {
	if queries == nil {
		return nil
	}
	pattern := "%" + term + "%"
	rows, err := queries.SearchUsers(context.Background(), pattern)
	if err != nil {
		log.Println("SearchUsers query failed:", err)
		return nil
	}
	users := make([]StlDevsUser, 0, len(rows))
	for _, row := range rows {
		user := &StlDevsUser{
			User: &github.User{
				Login:       stringPtr(row.Login),
				Name:        nullStringPtr(row.Name),
				Followers:   nullIntPtr(row.Followers),
				PublicRepos: nullIntPtr(row.PublicRepos),
				PublicGists: nullIntPtr(row.PublicGists),
				AvatarURL:   nullStringPtr(row.AvatarUrl),
				Type:        nullStringPtr(row.Type),
			},
			Hide:    row.Hide,
			IsAdmin: row.IsAdmin,
			Stars:   int(row.Stars),
			Forks:   int(row.Forks),
		}
		users = append(users, *user)
	}
	return users
}

var SearchRepos = func(term string) []stldevs.Repository {
	if queries == nil {
		return nil
	}
	pattern := "%" + term + "%"
	rows, err := queries.SearchRepos(context.Background(), pattern)
	if err != nil {
		log.Println("SearchRepos query failed:", err)
		return nil
	}
	return convertAggRepos(rows)
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

func userFromRow(row sqlc.GetUserRow) *StlDevsUser {
	company := row.Company
	ghUser := &github.User{
		Login:       stringPtr(row.Login),
		Email:       nullStringPtr(row.Email),
		Name:        nullStringPtr(row.Name),
		Location:    nullStringPtr(row.Location),
		Hireable:    nullBoolPtr(row.Hireable),
		Blog:        nullStringPtr(row.Blog),
		Bio:         nullStringPtr(row.Bio),
		Followers:   nullIntPtr(row.Followers),
		Following:   nullIntPtr(row.Following),
		PublicRepos: nullIntPtr(row.PublicRepos),
		PublicGists: nullIntPtr(row.PublicGists),
		AvatarURL:   nullStringPtr(row.AvatarUrl),
		Type:        nullStringPtr(row.Type),
		DiskUsage:   nullIntPtr(row.DiskUsage),
		CreatedAt:   githubTimestampPtr(row.CreatedAt),
		UpdatedAt:   githubTimestampPtr(row.UpdatedAt),
		Company:     &company,
	}
	return &StlDevsUser{
		User:    ghUser,
		Hide:    row.Hide,
		IsAdmin: row.IsAdmin,
	}
}

func convertAggRepos(rows []sqlc.AggRepo) []stldevs.Repository {
	repos := make([]stldevs.Repository, 0, len(rows))
	for _, row := range rows {
		repos = append(repos, repoFromAggRow(row))
	}
	return repos
}

func repoFromAggRow(row sqlc.AggRepo) stldevs.Repository {
	return stldevs.Repository{
		Owner:            stringPtr(row.Owner),
		Name:             stringPtr(row.Name),
		Description:      nullStringPtr(row.Description),
		Language:         nullStringPtr(row.Language),
		Homepage:         nullStringPtr(row.Homepage),
		ForksCount:       nullIntPtr(row.ForksCount),
		NetworkCount:     nullIntPtr(row.NetworkCount),
		OpenIssuesCount:  nullIntPtr(row.OpenIssuesCount),
		StargazersCount:  nullIntPtr(row.StargazersCount),
		SubscribersCount: nullIntPtr(row.SubscribersCount),
		WatchersCount:    nullIntPtr(row.WatchersCount),
		Size:             nullIntPtr(row.Size),
		Fork:             nullBoolPtr(row.Fork),
		DefaultBranch:    nullStringPtr(row.DefaultBranch),
		MasterBranch:     nullStringPtr(row.MasterBranch),
		CreatedAt:        nullTimePtr(row.CreatedAt),
		PushedAt:         nullTimePtr(row.PushedAt),
		UpdatedAt:        nullTimePtr(row.UpdatedAt),
		RefreshedAt:      nullTimePtr(row.RefreshedAt),
	}
}

func repoFromLanguageRow(row sqlc.LanguageLeadersRow) stldevs.Repository {
	return stldevs.Repository{
		Owner:           stringPtr(row.Owner),
		Name:            stringPtr(row.Name),
		Description:     nullStringPtr(row.Description),
		ForksCount:      nullIntPtr(row.ForksCount),
		StargazersCount: nullIntPtr(row.StargazersCount),
		WatchersCount:   nullIntPtr(row.WatchersCount),
		Fork:            nullBoolPtr(row.Fork),
	}
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	value := ns.String
	return &value
}

func nullStringValue(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func nullBoolPtr(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	value := nb.Bool
	return &value
}

func nullIntPtr(ni sql.NullInt32) *int {
	if !ni.Valid {
		return nil
	}
	value := int(ni.Int32)
	return &value
}

func nullTimePtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

func githubTimestampPtr(nt sql.NullTime) *github.Timestamp {
	if !nt.Valid {
		return nil
	}
	return &github.Timestamp{Time: nt.Time}
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func formatInt(ni sql.NullInt32) string {
	if !ni.Valid {
		return ""
	}
	return strconv.FormatInt(int64(ni.Int32), 10)
}
