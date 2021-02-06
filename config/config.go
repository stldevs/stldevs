package config

import (
	"encoding/json"
	"io"
)

type Config struct {
	GithubKey,
	Postgres,
	GithubClientID,
	GithubClientSecret,
	SessionSecret string
	Environment string
}

func NewConfig(r io.Reader) (*Config, error) {
	cfg := &Config{}
	err := json.NewDecoder(r).Decode(cfg)
	return cfg, err
}
