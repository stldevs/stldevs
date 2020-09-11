package aggregator

import (
	"context"
	"log"

	"github.com/google/go-github/v29/github"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type Aggregator struct {
	client  *github.Client
	db      *sqlx.DB
	running bool
}

func New(db *sqlx.DB, githubKey string) *Aggregator {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubKey})
	client := oauth2.NewClient(context.Background(), ts)
	return &Aggregator{db: db, client: github.NewClient(client)}
}

func (a *Aggregator) Run() {
	if a.running {
		log.Println("Already running, aborting run.")
		return
	}
	log.Println("Run started")
	a.running = true
	defer func() { a.running = false }()
	if err := a.insertRunLog(); err != nil {
		log.Println(err)
		return
	}
	log.Println("Run log inserted")
	users, err := FindInStl(a.client, "user")
	if err != nil {
		log.Println(err)
		return
	}
	orgs, err := FindInStl(a.client, "org")
	if err != nil {
		log.Println(err)
		return
	}
	for o := range orgs {
		users[o] = struct{}{}
	}
	for user := range users {
		log.Println("Adding/Updating", user)
		if err = a.Add(user); err != nil {
			log.Println(err)
			continue
		}
		log.Println("Updating repos of", user)
		_ = a.updateUsersRepos(user)
	}
}

func (a *Aggregator) Running() bool {
	return a.running
}
