package config

import (
	"os"
	"errors"
)

type Config struct {
	MysqlIp,
	MysqlPw,
	GithubKey,
	GithubClientID,
	GithubClientSecret,
	SessionSecret string
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	cfg.MysqlIp = os.Getenv("DB_PORT")
	cfg.MysqlPw = os.Getenv("mysqlpw")
	cfg.GithubKey = os.Getenv("github.key")
	cfg.GithubClientID = os.Getenv("github.client.id")
	cfg.GithubClientSecret = os.Getenv("github.client.secret")
	cfg.SessionSecret = os.Getenv("session.secret")
	return cfg, Exists(cfg.MysqlPw, cfg.GithubClientID, cfg.GithubClientSecret, cfg.GithubKey, cfg.SessionSecret)
}

func Exists(vals ...string) error {
	for _, val := range vals {
		if val == "" {
			return errors.New("Value is empty")
		}
	}
	return nil
}
