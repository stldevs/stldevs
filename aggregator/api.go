package aggregator

import (
	"database/sql"
	_ "embed"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/v52/github"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/db/sqlc"
)

//go:embed orgs.txt
var orgList string

type Aggregator struct {
	client  *github.Client
	queries *sqlc.Queries
	running bool
}

func NewScraperClient(cfg *config.Config) (*github.Client, error) {
	if cfg.ScraperClientID == "" || cfg.ScraperClientSecret == "" {
		return nil, errors.New("both ScraperClientID and ScraperClientSecret are required")
	}
	transport := &github.UnauthenticatedRateLimitedTransport{
		ClientID: cfg.ScraperClientID, ClientSecret: cfg.ScraperClientSecret,
	}
	client := transport.Client()
	// The transport reapplies the app secret on each hop.
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" || req.URL.Host != "api.github.com" || len(via) >= 10 {
			return http.ErrUseLastResponse
		}
		return nil
	}
	return github.NewClient(client), nil
}

func New(db *sql.DB, client *github.Client) *Aggregator {
	return &Aggregator{client: client, queries: sqlc.New(db)}
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
	for _, org := range strings.Split(orgList, "\n") {
		users[org] = struct{}{}
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
