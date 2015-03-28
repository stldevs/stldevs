package aggregator

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
)

func (agg *Aggregator) updateUsers(users map[string]struct{}) {
	// add users that have already been added previously to this list
	rows, err := agg.db.Query(`SELECT login
	FROM agg_user
	`)
	if err != nil {
		check(err)
	}
	for rows.Next() {
		user := github.User{}
		if err = rows.Scan(&user.Login); err != nil {
			check(err)
		}
		if _, ok := users[*user.Login]; !ok {
			users[*user.Login] = struct{}{}
		}
	}

	// update users in 10 worker pools, c is the input channel
	c := make(chan string)
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(c chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			for user := range c {
				agg.Add(user)
			}
		}(c, &wg)
	}

	// pump users through the input channel until depleted
	for user, _ := range users {
		c <- user
	}

	close(c)
	wg.Wait()
}

func (agg *Aggregator) updateRepos() {
	c := make(chan string)
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(c chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			for user := range c {
				agg.updateUsersRepos(user)
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

func (a *Aggregator) updateUsersRepos(user string) {
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

func (a *Aggregator) findStlUsers() map[string]struct{} {
	searchString := `location:"St. Louis"  location:"STL" location:"St Louis" location:"Saint Louis"`
	opts := &github.SearchOptions{Sort: "followers", Order: "desc", ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	users := map[string]struct{}{}
	for {
		result, resultResp, err := a.client.Search.Users(searchString, opts)
		check(err)
		checkRespAndWait(resultResp)
		for _, user := range result.Users {
			users[*user.Login] = struct{}{}
		}
		if resultResp.NextPage == 0 {
			break
		}

		opts.ListOptions.Page = resultResp.NextPage
	}
	fmt.Printf("Total found: %v\n", len(users))
	return users
}

func (a *Aggregator) Add(user string) {
	u, resp, err := a.client.Users.Get(user)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			log.Println("Deleting user", user)
			// remove user from database
			_, err = a.db.Exec(`DELETE FROM agg_user WHERE login=?`, user)
			check(err)
			return
		}
		log.Println(err)
		return
	}
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
