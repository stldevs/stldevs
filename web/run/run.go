package run

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/db"
)

var Routes = []crud.Spec{{
	Method:      "GET",
	Path:        "/runs",
	Handler:     List,
	Description: "Gets the time of the last scrape of GitHub",
	Tags:        []string{"Last Run"},
	Validate:    crud.Validate{},
}}

var epoch time.Time

func List(w http.ResponseWriter, r *http.Request) {
	if lastRun := db.LastRun(); lastRun.Year() == epoch.Year() {
		http.Error(w, "Failed to list", 500)
		return
	} else {
		jsonResponse(w, 200, lastRun)
	}
}

func jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
