package repo

import (
	"encoding/json"
	"net/http"

	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/db"
)

var Routes = []crud.Spec{{
	Method:      "GET",
	Path:        "/repos",
	Handler:     List,
	Description: "Lists repositories",
	Tags:        []string{"Repos"},
	Validate: crud.Validate{
		Query: crud.Object(map[string]crud.Field{
			"q": crud.String().Required().Description("Query string"),
		}),
	},
}}

func List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q is a required query parameter", 400)
		return
	}
	jsonResponse(w, 200, db.SearchRepos(q))
}

func jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
