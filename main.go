package main

import (
	"os"

	"encoding/json"

	"log"

	"github.com/jakecoffman/stldevs/web"
)

func main() {
	config := web.Config{}
	f, err := os.Open("config.json")
	if err != nil {
		log.Println("Couldn't find dev_config.json")
		return
	}

	json.NewDecoder(f).Decode(&config)

	if config.MysqlPw == "" || config.GithubKey == "" {
		log.Println("Config file missing important things")
		return
	}

	web.Run(config)
}
