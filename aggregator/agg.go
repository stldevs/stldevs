package aggregator

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"code.google.com/p/goauth2/oauth"

	"github.com/google/go-github/github"
)

type Aggregator struct {
	client *github.Client
	Users  []github.User
}

func NewAggregator() *Aggregator {
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: os.Getenv("GITHUB_API_KEY")},
	}

	var agg Aggregator
	agg.client = github.NewClient(t.Client())
	agg.Users = getUsers(agg.client)
	return &agg
}

func (a *Aggregator) GetRepos(user string) []github.Repository {
	opts := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc", ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	result, resp, err := a.client.Repositories.List(user, opts)
	check(err)
	checkRespAndWait(resp)
	return result
}

func getUsers(client *github.Client) []github.User {
	searchString := `location:"St. Louis"  location:"STL" location:"St Louis" location:"Saint Louis"`
	opts := &github.SearchOptions{Sort: "followers", Order: "desc", ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	result, resultResp, err := client.Search.Users(searchString, opts)
	check(err)
	checkRespAndWait(resultResp)
	// resp.NextPage
	fmt.Printf("Total found: %v\n", *result.Total)

	details := []github.User{}
	f, err := os.Open("users.json")
	if err != nil {
		for _, u := range result.Users {
			user, userResp, err := client.Users.Get(*u.Login)
			check(err)
			checkRespAndWait(userResp)
			details = append(details, *user)
		}
	} else {
		defer f.Close()
		// TODO: check for updates to users
		json.NewDecoder(f).Decode(&details)
	}

	f, err = os.Create("users.json")
	check(err)
	defer f.Close()
	json.NewEncoder(f).Encode(details)

	return details
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
