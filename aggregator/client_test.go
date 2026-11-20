package aggregator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jakecoffman/stldevs/config"
)

func TestScraperCredentials(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, secret, ok := r.BasicAuth()
		if !ok || id != "scraper" || secret != "test-secret" {
			t.Error("expected scraper app credentials")
		}
		_, _ = w.Write([]byte(`{"login":"octocat"}`))
	}))
	defer api.Close()
	cfg := &config.Config{GithubClientID: "website", GithubClientSecret: "login-secret"}
	if _, err := NewScraperClient(cfg); err == nil {
		t.Fatal("website credentials alone must not enable scraping")
	}
	cfg.ScraperClientID = "scraper"
	if _, err := NewScraperClient(cfg); err == nil {
		t.Fatal("incomplete scraper credentials accepted")
	}
	cfg.ScraperClientSecret = "test-secret"
	client, err := NewScraperClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL, _ = url.Parse(api.URL + "/")
	if _, _, err := client.Users.Get(context.Background(), "octocat"); err != nil {
		t.Fatal(err)
	}
}
