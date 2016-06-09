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
	_, err := db.Exec(createMeta)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(createUser)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(createRepo)
	if err != nil {
		log.Fatal(err)
	}

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
	users, err := a.findStlUsers()
	if err != nil {
		log.Println(err)
		return
	}
	if err = a.updateUsers(users); err != nil {
		log.Println(err)
		return
	}
	if err = a.updateRepos(); err != nil {
		log.Println(err)
	}
}

func (a *Aggregator) Running() bool {
	return a.running
}

