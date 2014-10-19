package main

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jakecoffman/stl-dev-stats/aggregator"
)

func main() {
	db, err := sql.Open("mysql", "root:bird@/github")
	check(err)
	defer db.Close()
	agg := aggregator.NewAggregator(db)

	step2(db, agg)
}

func step1(agg *aggregator.Aggregator) {
	users := agg.SearchUsers()

	for _, user := range users {
		fmt.Println("Getting user", *user.Login)
		agg.GatherUserDetails(*user.Login)
	}
}

func step2(db *sql.DB, agg *aggregator.Aggregator) {
	c := make(chan string)
	wg := sync.WaitGroup{}
	// spin up 100 workers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(c chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			for user := range c {
				agg.GatherRepos(user)
			}
		}(c, &wg)
	}

	// pump the users through the channel
	rows, err := db.Query("select login from user")
	check(err)
	defer rows.Close()
	for rows.Next() {
		var user string
		rows.Scan(&user)
		c <- user
	}

	// wait for them to finish
	close(c)
	wg.Wait()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
