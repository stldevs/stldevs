package aggregator

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
)

func step1(agg *Aggregator) {
	users := agg.searchUsers()

	c := make(chan string)
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(c chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			for user := range c {
				agg.gatherUserDetails(user)
			}
		}(c, &wg)
	}

	for _, user := range users {
		c <- *user.Login
	}

	close(c)
	wg.Wait()
}

func step2(agg *Aggregator) {
	c := make(chan string)
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(c chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			for user := range c {
				agg.gatherRepos(user)
			}
		}(c, &wg)
	}

	// pump the users through the channel
	rows, err := agg.db.Query("select login from agg_user")
	check(err)
	defer rows.Close()
	for rows.Next() {
		var user string
		rows.Scan(&user)
		c <- user
	}

	// wait for them to finish
	close(c)
	wg.Wait()
}

func (a *Aggregator) gatherRepos(user string) {
	opts := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		result, resp, err := a.client.Repositories.List(user, opts)
		check(err)
		checkRespAndWait(resp)
		for _, repo := range result {
			stmt, err := a.db.Prepare(`REPLACE INTO agg_repo VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
			check(err)
			var pushedAt *time.Time
			if repo.PushedAt != nil {
				pushedAt = &repo.PushedAt.Time
			}
			_, err = stmt.Exec(
				repo.Owner.Login,
				repo.Name,
				repo.Description,
				repo.Language,
				repo.Homepage,
				repo.ForksCount,
				repo.NetworkCount,
				repo.OpenIssuesCount,
				repo.StargazersCount,
				repo.SubscribersCount,
				repo.WatchersCount,
				repo.Size,
				repo.Fork,
				repo.DefaultBranch,
				repo.MasterBranch,
				repo.CreatedAt.Time,
				pushedAt,
				repo.UpdatedAt.Time)
			check(err)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
}

func (a *Aggregator) searchUsers() []github.User {
	searchString := `location:"St. Louis"  location:"STL" location:"St Louis" location:"Saint Louis"`
	opts := &github.SearchOptions{Sort: "followers", Order: "desc", ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	users := []github.User{}
	for {
		result, resultResp, err := a.client.Search.Users(searchString, opts)
		check(err)
		checkRespAndWait(resultResp)
		users = append(users, result.Users...)
		if resultResp.NextPage == 0 {
			break
		}

		opts.ListOptions.Page = resultResp.NextPage
	}
	fmt.Printf("Total found: %v\n", len(users))
	return users
}

func (a *Aggregator) gatherUserDetails(user string) {
	u, resp, err := a.client.Users.Get(user)
	check(err)
	checkRespAndWait(resp)
	stmt, err := a.db.Prepare(`REPLACE INTO agg_user VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	check(err)
	_, err = stmt.Exec(
		u.Login,
		u.Email,
		u.Name,
		u.Location,
		u.Hireable,
		u.Blog,
		u.Bio,
		u.Followers,
		u.Following,
		u.PublicRepos,
		u.PublicGists,
		u.AvatarURL,
		u.DiskUsage,
		u.CreatedAt.Time,
		u.UpdatedAt.Time)
	check(err)
	stmt.Close()
}

func (a *Aggregator) insertRunLog() {
	stmt, err := a.db.Prepare(`INSERT INTO agg_meta VALUES (?)`)
	check(err)

	_, err = stmt.Exec(time.Now())
	check(err)
}

func checkRespAndWait(r *github.Response) {
	if r.Remaining == 0 {
		duration := time.Now().Sub(r.Rate.Reset.Time)
		fmt.Println("I ran out of requests, waiting", duration)
		time.Sleep(duration)
	} else {
		fmt.Println(r.Remaining, "calls remaining until", r.Rate.Reset.String())
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		debug.PrintStack()
		os.Exit(1)
	}
}
