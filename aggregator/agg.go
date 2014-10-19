package aggregator

import (
	"database/sql"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"code.google.com/p/goauth2/oauth"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
)

type Aggregator struct {
	client *github.Client
	db     *sql.DB
}

func NewAggregator(db *sql.DB) *Aggregator {
	// init db
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user (
			login VARCHAR(255) NOT NULL PRIMARY KEY,
			email TEXT,
			name TEXT,
			location TEXT,
			hireable BOOL,
			blog TEXT,
			bio TEXT,
			followers INTEGER,
			following INTEGER,
			public_repos INTEGER,
			public_gists INTEGER,
			avatar_url TEXT,
			disk_usage INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		);`)
	check(err)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS repo (
			owner VARCHAR(255),
			name TEXT,
			description TEXT,
			language TEXT,
			homepage TEXT,
			forks_count INT,
			network_count INT,
			open_issues_count INT,
			stargazers_count INT,
			subscribers_count INT,
			watchers_count INT,
			size INT,
			fork BOOL,
			default_branch TEXT,
			master_branch TEXT,
			created_at DATETIME,
			pushed_at DATETIME,
			updated_at DATETIME
		);`)
	check(err)

	// init client
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: os.Getenv("GITHUB_API_KEY")},
	}

	var agg Aggregator
	agg.db = db
	agg.client = github.NewClient(t.Client())
	return &agg
}

func (a *Aggregator) GatherRepos(user string) {
	opts := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		result, resp, err := a.client.Repositories.List(user, opts)
		check(err)
		checkRespAndWait(resp)
		for _, repo := range result {
			stmt, err := a.db.Prepare(`REPLACE INTO repo VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
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

func (a *Aggregator) SearchUsers() []github.User {
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

func (a *Aggregator) GatherUserDetails(user string) {
	u, resp, err := a.client.Users.Get(user)
	check(err)
	checkRespAndWait(resp)
	stmt, err := a.db.Prepare(`REPLACE INTO user VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
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
