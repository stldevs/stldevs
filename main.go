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

	setupLogger("log.txt")
	cfg, err := parseConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}
	web.Run(cfg)
}

func setupLogger(fileName string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, file))
}

func parseConfig(fileName string) (*config.Config, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	cfg := &config.Config{}
	err = json.NewDecoder(f).Decode(cfg)
	return cfg, err
}
