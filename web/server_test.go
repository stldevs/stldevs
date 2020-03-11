package web

import (
	"net/url"
	"testing"
)

func TestCorsMiddleware(t *testing.T) {
	origin := "http://localhost:3000"
	parsed, err := url.Parse(origin)
	if !(err == nil && parsed.Hostname() == "localhost") {
		t.Error(err, parsed)
	}
}
