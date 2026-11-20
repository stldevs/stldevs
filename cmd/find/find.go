package main

import (
	"fmt"
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

	client, err := aggregator.NewScraperClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	u, err := aggregator.FindInStl(client, "user")
	if err != nil {
		log.Fatal(err)
	}
	for k := range u {
		fmt.Println(k)
	}
}
