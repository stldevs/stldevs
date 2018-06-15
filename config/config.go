package config

import (
	"encoding/json"
	"io"
)

type Config struct {
	GithubKey,
	MysqlPw,
	PostgresPw,
	GithubClientID,
	GithubClientSecret,
	SessionSecret string
}

func NewConfig(r io.Reader) (*Config, error) {
	cfg := &Config{}
	err := json.NewDecoder(r).Decode(cfg)
	return cfg, err
}
