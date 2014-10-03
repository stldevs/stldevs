package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"code.google.com/p/goauth2/oauth"

	"github.com/google/go-github/github"
)

func main() {
	searchString := `location:"St. Louis"  location:"STL" location:"St Louis" location:"Saint Louis"`

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: "731fc12b627c1739c6d610f453a13a9cf9aebe2e"},
	}

	client := github.NewClient(t.Client())
	opts := &github.SearchOptions{Sort: "repositories", Order: "desc", ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	result, resultResp, err := client.Search.Users(searchString, opts)
	check(err)
	checkRespAndWait(resultResp)
	fmt.Println(resultResp.Remaining, "searches remaining")
	// resp.NextPage
	fmt.Printf("Total found: %v\n", *result.Total)
	f, err := os.Create("out.csv")
	defer f.Close()
	check(err)
	writer := csv.NewWriter(f)
	writer.Write([]string{
		"login",
		"email",
		"blog",
		"company",
		"public_repos",
		"created_at",
	})

	for _, user := range result.Users {
		user, userResp, err := client.Users.Get(*user.Login)
		check(err)
		checkRespAndWait(userResp)
		fmt.Println(userResp.Remaining, "user gets remaining")
		writer.Write([]string{
			get(user.Login),
			get(user.Email),
			get(user.Blog),
			get(user.Company),
			getI(user.PublicRepos),
			user.CreatedAt.String(),
		})
		writer.Flush()
	}
}

func checkRespAndWait(r *github.Response) {
	if r.Remaining == 0 {
		duration := time.Now().Sub(r.Rate.Reset.Time)
		fmt.Println("I ran out of requests, waiting", duration)
		time.Sleep(duration)
	}
}

// safely dereference for output
func get(s *string) string {
	if s != nil {
		return *s
	} else {
		return ""
	}
}

// safely dereference for output
func getI(s *int) string {
	if s != nil {
		return strconv.Itoa(*s)
	} else {
		return ""
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		debug.PrintStack()
		os.Exit(1)
	}
}
