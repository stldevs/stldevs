package main

import (
	"encoding/json"
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
		Token: &oauth.Token{AccessToken: os.Getenv("GITHUB_API_KEY")},
	}

	client := github.NewClient(t.Client())
	opts := &github.SearchOptions{Sort: "followers", Order: "desc", ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	result, resultResp, err := client.Search.Users(searchString, opts)
	check(err)
	checkRespAndWait(resultResp)
	// resp.NextPage
	fmt.Printf("Total found: %v\n", *result.Total)

	records := []map[string]string{}
	for _, user := range result.Users {
		user, userResp, err := client.Users.Get(*user.Login)
		check(err)
		checkRespAndWait(userResp)

		record := map[string]string{
			"login":        get(user.Login),
			"email":        get(user.Email),
			"blog":         get(user.Blog),
			"company":      get(user.Company),
			"public_repos": getI(user.PublicRepos),
			"followers":    getI(user.Followers),
			"created_at":   user.CreatedAt.String(),
		}
		records = append(records, record)
	}

	f2, err := os.Create("out.json")
	check(err)
	defer f2.Close()
	json.NewEncoder(f2).Encode(records)

	// f, err := os.Create("out.csv")
	// check(err)
	// defer f.Close()
	// writer := csv.NewWriter(f)
	// writer.Write([]string{
	// 	"login",
	// 	"email",
	// 	"blog",
	// 	"company",
	// 	"public_repos",
	// 	"followers",
	// 	"created_at",
	// })

	// writer.WriteAll(records)
	// writer.Flush()
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
