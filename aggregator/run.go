package aggregator

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v52/github"
	"github.com/jakecoffman/stldevs/db/sqlc"
)

func (a *Aggregator) updateUsersRepos(user string) error {
	ctx := context.Background()
	now := time.Now()

	opts := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		result, resp, err := a.client.Repositories.List(ctx, user, opts)
		if shouldTryAgain(resp) {
			continue
		}
		if err != nil {
			log.Println(err)
			return err
		}
		for _, repo := range result {
			params, err := buildRepoParams(repo, now)
			if err != nil {
				log.Println(err)
				continue
			}
			updated, err := a.queries.UpdateRepo(ctx, toUpdateRepoParams(params))
			if err != nil {
				log.Println(err)
				return err
			}
			if updated == 0 {
				if err := a.queries.InsertRepo(ctx, params); err != nil {
					log.Println("Error executing replace into agg_repo", err)
					return err
				}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	deleted, err := a.queries.DeleteReposByOwnerBefore(ctx, sqlc.DeleteReposByOwnerBeforeParams{
		Owner:       user,
		RefreshedAt: sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		log.Printf("Error deleting out of date repos for user %v: %v", user, err)
		return err
	}
	log.Printf("Deleted %v repos that user %v was missing", deleted, user)
	return nil
}

func FindInStl(client *github.Client, typ string) (map[string]struct{}, error) {
	users := map[string]struct{}{}

	// since github limits to 1000 results, break the search up with created
	for _, date := range []string{
		"created:<2013-01-01",
		"created:2013-01-01..2014-01-01",
		"created:2014-01-01..2015-01-01",
		"created:2015-01-01..2016-01-01",
		"created:2016-01-01..2017-01-01",
		"created:2017-01-01..2018-01-01",
		"created:2018-01-01..2019-01-01",
		"created:2019-01-01..2020-01-01",
		"created:2020-01-01..2021-01-01",
		"created:2021-01-01..2022-01-01",
		"created:2022-01-01..2023-01-01",
		"created:2023-01-01..2024-01-01",
		"created:2024-01-01..2025-01-01",
		"created:>2025-01-01",
	} {
		const locations = `location:"St. Louis" location:"STL" location:"St Louis" location:"Saint Louis"`
		searchString := fmt.Sprintf(`%v %v repos:>1 type:"%v"`, locations, date, typ)
		opts := &github.SearchOptions{
			ListOptions: github.ListOptions{Page: 1, PerPage: 100},
			Sort:        "repositories",
		}
		for {
			time.Sleep(2 * time.Second)
			result, resultResp, err := client.Search.Users(context.Background(), searchString, opts)
			if shouldTryAgain(resultResp) {
				continue
			}
			if err != nil {
				log.Println(err)
				return users, err
			}
			for _, user := range result.Users {
				users[*user.Login] = struct{}{}
			}
			if resultResp.NextPage == 10 {
				log.Println("Warning: hit the limit on user search for date", date)
			}
			if resultResp.NextPage == 0 {
				break
			}

			opts.ListOptions.Page = resultResp.NextPage
		}
	}
	fmt.Printf("total of type %v found: %v\n", typ, len(users))
	return users, nil
}

func (a *Aggregator) Add(user string) error {
start:
	u, resp, err := a.client.Users.Get(context.Background(), user)
	if shouldTryAgain(resp) {
		goto start
	}
	if err != nil || u == nil {
		log.Println("Failed getting user details for", user, ":", err)
		return err
	}
	updateParams, insertParams, err := buildUserParams(u, time.Now())
	if err != nil {
		log.Println(err)
		return err
	}
	updated, err := a.queries.UpdateUser(context.Background(), updateParams)
	if err != nil {
		log.Println(err)
		return err
	}
	if updated == 0 {
		if err := a.queries.InsertUser(context.Background(), insertParams); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func (a *Aggregator) insertRunLog() error {
	if err := a.queries.InsertRunLog(context.Background(), time.Now()); err != nil {
		log.Println("Error executing insert", err)
		return err
	}
	return nil
}

func shouldTryAgain(r *github.Response) bool {
	if r.Rate.Remaining <= 0 {
		duration := time.Until(r.Rate.Reset.Time)
		fmt.Printf("I ran out of requests (%v), waiting %v\n", r.Rate.Limit, duration)
		time.Sleep(duration + time.Second)
		return true
	}
	return false
}

func buildRepoParams(repo *github.Repository, refreshedAt time.Time) (sqlc.InsertRepoParams, error) {
	if repo.Owner == nil || repo.Owner.Login == nil || *repo.Owner.Login == "" {
		return sqlc.InsertRepoParams{}, fmt.Errorf("repo missing owner")
	}
	if repo.Name == nil || *repo.Name == "" {
		return sqlc.InsertRepoParams{}, fmt.Errorf("repo missing name")
	}
	return sqlc.InsertRepoParams{
		Owner:            repo.GetOwner().GetLogin(),
		Name:             repo.GetName(),
		Description:      nullStringFromPtr(repo.Description),
		Language:         nullStringFromPtr(repo.Language),
		Homepage:         nullStringFromPtr(repo.Homepage),
		ForksCount:       nullInt32FromPtr(repo.ForksCount),
		NetworkCount:     nullInt32FromPtr(repo.NetworkCount),
		OpenIssuesCount:  nullInt32FromPtr(repo.OpenIssuesCount),
		StargazersCount:  nullInt32FromPtr(repo.StargazersCount),
		SubscribersCount: nullInt32FromPtr(repo.SubscribersCount),
		WatchersCount:    nullInt32FromPtr(repo.WatchersCount),
		Size:             nullInt32FromPtr(repo.Size),
		Fork:             nullBoolFromPtr(repo.Fork),
		DefaultBranch:    nullStringFromPtr(repo.DefaultBranch),
		MasterBranch:     nullStringFromPtr(repo.MasterBranch),
		CreatedAt:        nullTimeFromTimestamp(repo.CreatedAt),
		PushedAt:         nullTimeFromTimestamp(repo.PushedAt),
		UpdatedAt:        nullTimeFromTimestamp(repo.UpdatedAt),
		RefreshedAt:      sql.NullTime{Time: refreshedAt, Valid: true},
	}, nil
}

func toUpdateRepoParams(p sqlc.InsertRepoParams) sqlc.UpdateRepoParams {
	return sqlc.UpdateRepoParams{
		Owner:            p.Owner,
		Name:             p.Name,
		Description:      p.Description,
		Language:         p.Language,
		Homepage:         p.Homepage,
		ForksCount:       p.ForksCount,
		NetworkCount:     p.NetworkCount,
		OpenIssuesCount:  p.OpenIssuesCount,
		StargazersCount:  p.StargazersCount,
		SubscribersCount: p.SubscribersCount,
		WatchersCount:    p.WatchersCount,
		Size:             p.Size,
		Fork:             p.Fork,
		DefaultBranch:    p.DefaultBranch,
		MasterBranch:     p.MasterBranch,
		CreatedAt:        p.CreatedAt,
		PushedAt:         p.PushedAt,
		UpdatedAt:        p.UpdatedAt,
		RefreshedAt:      p.RefreshedAt,
	}
}

func buildUserParams(u *github.User, refreshedAt time.Time) (sqlc.UpdateUserParams, sqlc.InsertUserParams, error) {
	if u.Login == nil || *u.Login == "" {
		return sqlc.UpdateUserParams{}, sqlc.InsertUserParams{}, fmt.Errorf("user missing login")
	}
	update := sqlc.UpdateUserParams{
		Login:       u.GetLogin(),
		Email:       nullStringFromPtr(u.Email),
		Name:        nullStringFromPtr(u.Name),
		Location:    nullStringFromPtr(u.Location),
		Hireable:    nullBoolFromPtr(u.Hireable),
		Blog:        nullStringFromPtr(u.Blog),
		Bio:         nullStringFromPtr(u.Bio),
		Followers:   nullInt32FromPtr(u.Followers),
		Following:   nullInt32FromPtr(u.Following),
		PublicRepos: nullInt32FromPtr(u.PublicRepos),
		PublicGists: nullInt32FromPtr(u.PublicGists),
		AvatarUrl:   nullStringFromPtr(u.AvatarURL),
		Type:        nullStringFromPtr(u.Type),
		DiskUsage:   nullInt32FromPtr(u.DiskUsage),
		CreatedAt:   nullTimeFromTimestamp(u.CreatedAt),
		UpdatedAt:   nullTimeFromTimestamp(u.UpdatedAt),
		RefreshedAt: sql.NullTime{Time: refreshedAt, Valid: true},
		Company:     u.GetCompany(),
	}
	insert := sqlc.InsertUserParams{
		Login:       update.Login,
		Email:       update.Email,
		Name:        update.Name,
		Location:    update.Location,
		Hireable:    update.Hireable,
		Blog:        update.Blog,
		Bio:         update.Bio,
		Followers:   update.Followers,
		Following:   update.Following,
		PublicRepos: update.PublicRepos,
		PublicGists: update.PublicGists,
		AvatarUrl:   update.AvatarUrl,
		Type:        update.Type,
		DiskUsage:   update.DiskUsage,
		CreatedAt:   update.CreatedAt,
		UpdatedAt:   update.UpdatedAt,
		RefreshedAt: update.RefreshedAt,
		Company:     update.Company,
	}
	return update, insert, nil
}

func nullStringFromPtr(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func nullInt32FromPtr(value *int) sql.NullInt32 {
	if value == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*value), Valid: true}
}

func nullBoolFromPtr(value *bool) sql.NullBool {
	if value == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *value, Valid: true}
}

func nullTimeFromTimestamp(ts *github.Timestamp) sql.NullTime {
	if ts == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: ts.Time, Valid: true}
}
