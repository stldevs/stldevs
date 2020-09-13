package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"log"
	"os"

	"github.com/jakecoffman/stldevs/aggregator"
	"github.com/jakecoffman/stldevs/config"
)

func main() {
	f, err := os.Open("./config.json") // TODO: make configurable
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.NewConfig(f)
	if err != nil {
		log.Fatal(err)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.GithubKey})
	httpClient := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(httpClient)

	u, _ := aggregator.FindInStl(client, "user")
	for k := range u {
		fmt.Println(k)
	}
}
