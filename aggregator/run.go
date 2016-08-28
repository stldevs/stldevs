package aggregator

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
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
			_, err = a.db.Exec(`DELETE FROM agg_user WHERE login=?`, *existing.Login)
			if err != nil {
				log.Println("Error while deleting moved user:", *existing.Login, err)
			}
			_, err = a.db.Exec(`DELETE FROM agg_repo WHERE owner=?`, *existing.Login)
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
		result, resp, err := a.client.Repositories.List(user, opts)
		if checkRespAndWait(resp, err) != nil {
			log.Println(err)
			return err
		}
		for _, repo := range result {
			var pushedAt *time.Time
			if repo.PushedAt != nil {
				pushedAt = &repo.PushedAt.Time
			}
			_, err = a.db.Exec(`REPLACE INTO agg_repo VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				repo.Owner.Login, repo.Name, repo.Description, repo.Language, repo.Homepage,
				repo.ForksCount, repo.NetworkCount, repo.OpenIssuesCount, repo.StargazersCount, repo.SubscribersCount,
				repo.WatchersCount, repo.Size, *repo.Fork, repo.DefaultBranch, repo.MasterBranch, repo.CreatedAt.Time,
				pushedAt, repo.UpdatedAt.Time)
			if err != nil {
				log.Println("Error executing replace into agg_repo", err)
				return err
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
		result, resultResp, err := a.client.Search.Users(searchString, opts)
		if checkRespAndWait(resultResp, err) != nil {
			log.Println(err)
			return
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
	u, resp, err := a.client.Users.Get(user)
	if checkRespAndWait(resp, err) != nil {
		log.Println("Failed getting user details for", user, ":", err)
		return
	}
	_, err = a.db.Exec(`REPLACE INTO agg_user VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		u.Login, u.Email, u.Name, u.Location, u.Hireable, u.Blog, u.Bio, u.Followers, u.Following,
		u.PublicRepos, u.PublicGists, u.AvatarURL, u.Type, u.DiskUsage, u.CreatedAt.Time, u.UpdatedAt.Time)
	return err
}

func (a *Aggregator) insertRunLog() error {
	_, err := a.db.Exec(`INSERT INTO agg_meta VALUES (?)`, time.Now())
	if err != nil {
		log.Println("Error executing insert", err)
	}
	return err
}

func checkRespAndWait(r *github.Response, err error) error {
	if r.Remaining == 0 {
		duration := time.Now().Sub(r.Rate.Reset.Time)
		fmt.Println("I ran out of requests, waiting", duration)
		time.Sleep(duration)
	} else if err != nil {
		return err
	}
	return nil
}
