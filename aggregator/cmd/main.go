package main

import (
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jakecoffman/stldevs/aggregator"
)

func main() {
	db, err := sql.Open("mysql", "root:bird@/github")
	check(err)
	defer db.Close()
	agg := aggregator.NewAggregator(db)

	// step1(agg)
	step2(db, agg)
}

func step1(agg *aggregator.Aggregator) {
	users := agg.SearchUsers()

	c := make(chan string)
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(c chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			for user := range c {
				agg.GatherUserDetails(user)
			}
		}(c, &wg)
	}

	for _, user := range users {
		c <- *user.Login
	}

	close(c)
	wg.Wait()
}

func step2(db *sql.DB, agg *aggregator.Aggregator) {
	c := make(chan string)
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
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
