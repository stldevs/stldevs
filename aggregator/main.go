package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"code.google.com/p/goauth2/oauth"

	"github.com/google/go-github/github"
)

func main() {
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: os.Getenv("GITHUB_API_KEY")},
	}

	client := github.NewClient(t.Client())
	users := getUsers(client)
	details := getUserDetails(client, users)
	getRepos(client, details)
}

func getRepos(client *github.Client, users []github.User) {
	for _, u := range users {
		opts := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc", ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
		result, resp, err := client.Repositories.List(*u.Login, opts)
		check(err)
		checkRespAndWait(resp)
		f, err := os.Create(*u.Login + ".json")
		check(err)
		json.NewEncoder(f).Encode(result)
		f.Close()
	}
}

func getUserDetails(client *github.Client, users []github.User) []github.User {
	infile, err := os.Open("users.json")
	details := make([]github.User, 0)
	// TODO: check each user against the cache to see if they need to update
	if err == nil {
		json.NewDecoder(infile).Decode(&details)
		return details
	}
	fmt.Println("Unable to read from users cache:", err)

	for _, u := range users {
		user, userResp, err := client.Users.Get(*u.Login)
		check(err)
		checkRespAndWait(userResp)
		details = append(details, *user)
	}

	f, err := os.Create("users.json")
	check(err)
	json.NewEncoder(f).Encode(details)
	return details
}

func getUsers(client *github.Client) []github.User {
	searchString := `location:"St. Louis"  location:"STL" location:"St Louis" location:"Saint Louis"`
	opts := &github.SearchOptions{Sort: "followers", Order: "desc", ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	result, resultResp, err := client.Search.Users(searchString, opts)
	check(err)
	checkRespAndWait(resultResp)
	// resp.NextPage
	fmt.Printf("Total found: %v\n", *result.Total)
	return result.Users
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
