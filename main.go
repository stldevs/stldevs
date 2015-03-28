package main

import (
	"encoding/json"
	"log"
	"os"

	"runtime"

	"io"

	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/web"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, file))

	config := config.Config{}
	f, err := os.Open("config.json")
	if err != nil {
		log.Println("Couldn't find dev_config.json")
		return
	}

	json.NewDecoder(f).Decode(&config)

	web.Run(config)
}
