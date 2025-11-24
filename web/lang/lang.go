package lang

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/db"
)

var Routes = []crud.Spec{{
	Method:      "GET",
	Path:        "/langs",
	Handler:     List,
	Description: "List languages",
	Tags:        []string{"Languages"},
	Validate:    crud.Validate{},
}, {
	Method:      "GET",
	Path:        "/langs/{lang}",
	Handler:     Get,
	Description: "Gets a language and displays repo information",
	Tags:        []string{"Languages"},
	Validate: crud.Validate{
		Query: crud.Object(map[string]crud.Field{
			"limit":  crud.Number().Min(1).Max(25).Description("Maximum number of items to return"),
			"offset": crud.Number().Min(0).Description("Number of entries to skip"),
		}),
		Path: crud.Object(map[string]crud.Field{
			"lang": crud.String().Required().Description("The language name"),
		}),
	},
}}

func List(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, db.PopularLanguages())
}

func Get(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}

	lang := r.PathValue("lang")
	langs := db.Language(lang)

	if limit+offset > len(langs) {
		limit = len(langs)
	} else {
		limit += offset
	}
	if offset > len(langs) {
		limit = 0
		offset = 0
	}
	jsonResponse(w, 200, map[string]interface{}{
		"languages": langs[offset:limit],
		"count":     len(langs),
		"language":  lang,
	})
}

func jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
