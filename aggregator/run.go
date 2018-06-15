package aggregator

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
	"context"
)

func (a *Aggregator) removeUsersNotFoundInSearch(users map[string]struct{}) error {
	existingUsers := []github.User{}
	err := a.db.Select(&existingUsers, `SELECT login FROM agg_user`)
	if err != nil {
		log.Println("Error querying agg_user", err)
		return err
	}

	// remove users that no longer come up in search
	for _, existing := range existingUsers {
		if _, ok := users[*existing.Login]; !ok {
			log.Println(*existing.Login, "is no longer in St. Louis")
			_, err = a.db.Exec(`DELETE FROM agg_user WHERE login=$1`, *existing.Login)
			if err != nil {
				log.Println("Error while deleting moved user:", *existing.Login, err)
			}
			_, err = a.db.Exec(`DELETE FROM agg_repo WHERE owner=$1`, *existing.Login)
			if err != nil {
				log.Println("Error while deleting moved user's repos", *existing.Login, err)
			}
		}
	}
	return nil
}

func (a *Aggregator) updateUsersRepos(user string) error {
	opts := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		result, resp, err := a.client.Repositories.List(context.Background(), user, opts)
		if checkRespAndWait(resp, err) != nil {
			log.Println(err)
			return err
		}
		for _, repo := range result {
			var pushedAt *time.Time
			if repo.PushedAt != nil {
				pushedAt = &repo.PushedAt.Time
			}
			r, err := a.db.Exec(`UPDATE agg_repo
set owner = $1,
	name = $2,
	description = $3,
	language = $4,
	homepage = $5,
	forks_count = $6,
	network_count = $7,
	open_issues_count = $8,
	stargazers_count = $9,
	subscribers_count = $10,
	watchers_count = $11,
	size = $12,
	fork = $13,
	default_branch = $14,
	master_branch = $15,
	created_at = $16,
	pushed_at=$17,
	updated_at = $18
where owner=$1 and name=$2`, repo.Owner.Login, repo.Name, repo.Description, repo.Language, repo.Homepage,
				repo.ForksCount, repo.NetworkCount, repo.OpenIssuesCount, repo.StargazersCount, repo.SubscribersCount,
				repo.WatchersCount, repo.Size, *repo.Fork, repo.DefaultBranch, repo.MasterBranch, repo.CreatedAt.Time,
				pushedAt, repo.UpdatedAt.Time)
			if err != nil {
				log.Println(err)
				return err
			}
			if n, err := r.RowsAffected(); err != nil {
				log.Println(err)
				return err
			} else if n == 0 {
				_, err = a.db.Exec(`INSERT INTO agg_repo VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
					repo.Owner.Login, repo.Name, repo.Description, repo.Language, repo.Homepage,
					repo.ForksCount, repo.NetworkCount, repo.OpenIssuesCount, repo.StargazersCount, repo.SubscribersCount,
					repo.WatchersCount, repo.Size, *repo.Fork, repo.DefaultBranch, repo.MasterBranch, repo.CreatedAt.Time,
					pushedAt, repo.UpdatedAt.Time)
				if err != nil {
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
	return nil
}

func (a *Aggregator) FindInStl(typ string) (map[string]struct{}, error) {
	searchString := fmt.Sprintf(`location:"St. Louis" location:"STL" location:"St Louis" location:"Saint Louis" type:"%v"`, typ)
	opts := &github.SearchOptions{ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	users := map[string]struct{}{}
	for {
		result, resultResp, err := a.client.Search.Users(context.Background(), searchString, opts)
		if checkRespAndWait(resultResp, err) != nil {
			log.Println(err)
			return users, err
		}
		for _, user := range result.Users {
			users[*user.Login] = struct{}{}
		}
		if resultResp.NextPage == 0 {
			break
		}

		opts.ListOptions.Page = resultResp.NextPage
	}
	fmt.Printf("total devs in St. Louis found: %v\n", len(users))
	return users, nil
}

func (a *Aggregator) Add(user string) error {
	u, resp, err := a.client.Users.Get(context.Background(), user)
	if checkRespAndWait(resp, err) != nil || u == nil {
		log.Println("Failed getting user details for", user, ":", err)
		return err
	}
	r, err := a.db.Exec(`UPDATE agg_user 
set login = $1,
	email = $2,
	name = $3,
	location = $4,
	hireable = $5,
	blog = $6,
	bio = $7,
	followers = $8,
	following = $9,
	public_repos = $10,
	public_gists = $11,
	avatar_url = $12,
	type = $13,
	disk_usage = $14,
	created_at = $15,
	updated_at = $16
where login=$1`, u.Login, u.Email, u.Name, u.Location, u.Hireable, u.Blog, u.Bio, u.Followers, u.Following,
		u.PublicRepos, u.PublicGists, u.AvatarURL, u.Type, u.DiskUsage, u.CreatedAt.Time, u.UpdatedAt.Time)
	if err != nil {
		log.Println(err)
		return err
	}
	if n, err := r.RowsAffected(); err != nil {
		log.Println(err)
		return err
	} else if n == 0 {
		_, err = a.db.Exec(`INSERT INTO agg_user (
login, email, name, location, hireable, blog, bio, followers, following, public_repos, public_gists, avatar_url, type, disk_usage, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`, u.Login, u.Email, u.Name, u.Location, u.Hireable,
			u.Blog, u.Bio, u.Followers, u.Following, u.PublicRepos, u.PublicGists, u.AvatarURL, u.Type, u.DiskUsage,
			u.CreatedAt.Time, u.UpdatedAt.Time)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return err
}

func (a *Aggregator) insertRunLog() error {
	_, err := a.db.Exec(`INSERT INTO agg_meta VALUES ($1)`, time.Now())
	if err != nil {
		log.Println("Error executing insert", err)
	}
	return err
}

func checkRespAndWait(r *github.Response, err error) error {
	if r.Remaining == 0 {
		duration := r.Rate.Reset.Time.Sub(time.Now())
		fmt.Println("I ran out of requests, waiting", duration)
		time.Sleep(duration)
	} else if err != nil {
		return err
	}
	return nil
}
