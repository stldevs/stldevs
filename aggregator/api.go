package aggregator

import (
	"log"

	"golang.org/x/oauth2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
	"github.com/jmoiron/sqlx"
)

type Aggregator struct {
	client  *github.Client
	db      *sqlx.DB
	running bool
}

func New(db *sqlx.DB, githubKey string) *Aggregator {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubKey})
	client := oauth2.NewClient(oauth2.NoContext, ts)
	return &Aggregator{db: db, client: github.NewClient(client)}
}

func (a *Aggregator) Run() {
	if a.running {
		return
	}
	a.running = true
	defer func() { a.running = false }()
	if err := a.insertRunLog(); err != nil {
		log.Println(err)
		return
	}
	users, err := a.FindInStl("user")
	if err != nil {
		log.Println(err)
		return
	}
	orgs, err := a.FindInStl("org")
	if err != nil {
		log.Println(err)
		return
	}
	for o := range orgs {
		users[o] = struct{}{}
	}
	if err = a.removeUsersNotFoundInSearch(users); err != nil {
		return
	}
	for user := range users {
		log.Println("Adding/Updating", user)
		if err = a.Add(user); err != nil {
			continue
		}
		log.Println("Updating repos of", user)
		a.updateUsersRepos(user)
	}
}

func (a *Aggregator) Running() bool {
	return a.running
}

