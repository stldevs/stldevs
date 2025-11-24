package dev

import (
	"encoding/json"
	"net/http"

	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/sessions"
	"github.com/jakecoffman/stldevs/web/auth"
)

var Routes = []crud.Spec{{
	Method:      "GET",
	Path:        "/devs",
	Handler:     List,
	Description: "List devs",
	Tags:        []string{"Devs"},
	Validate: crud.Validate{
		Query: crud.Object(map[string]crud.Field{
			"q":       crud.String().Description("Search query"),
			"type":    crud.String().Description("Type of dev"),
			"company": crud.String().Description("Company"),
		}),
	},
}, {
	Method:      "GET",
	Path:        "/devs/{login}",
	Handler:     Get,
	Description: "Get a dev profile",
	Tags:        []string{"Devs"},
	Validate: crud.Validate{
		Path: crud.Object(map[string]crud.Field{
			"login": crud.String().Required().Description("GitHub login"),
		}),
	},
}, {
	Method:      "PATCH",
	Path:        "/devs/{login}",
	PreHandlers: auth.Authenticated,
	Handler:     Patch,
	Description: "Update a dev profile",
	Tags:        []string{"Devs"},
	Validate: crud.Validate{
		Path: crud.Object(map[string]crud.Field{
			"login": crud.String().Required().Description("GitHub login"),
		}),
		Body: crud.Object(map[string]crud.Field{
			"hide": crud.Boolean().Required(),
		}),
	},
}, {
	Method:      "DELETE",
	Path:        "/devs/{login}",
	PreHandlers: auth.Authenticated,
	Handler:     Delete,
	Description: "Delete a dev profile",
	Tags:        []string{"Devs"},
	Validate: crud.Validate{
		Path: crud.Object(map[string]crud.Field{
			"login": crud.String().Required().Description("GitHub login"),
		}),
	},
}}

type ListQuery struct {
	Q       string `form:"q"`
	Type    string `form:"type"`
	Company string `form:"company"`
}

func List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	typ := r.URL.Query().Get("type")
	company := r.URL.Query().Get("company")

	if (q == "" && typ == "") || (q != "" && typ != "") {
		http.Error(w, "provide either the type query parameter or the q query parameter", 400)
		return
	}

	if q != "" {
		jsonResponse(w, 200, db.SearchUsers(q))
		return
	}

	if listing := db.PopularDevs(typ, company); listing == nil {
		http.Error(w, "Failed to list", 500)
	} else {
		jsonResponse(w, 200, listing)
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	profile, err := db.Profile(r.PathValue("login"))
	if err != nil {
		http.Error(w, "Failed to find user", 404)
		return
	}
	jsonResponse(w, 200, profile)
}

type UpdateUser struct {
	Hide bool `json:"hide"`
}

// Patch allows users and admins show or hide themselves in the site
func Patch(w http.ResponseWriter, r *http.Request) {
	login := r.PathValue("login")
	session := sessions.GetEntry(r)
	if session.User.IsAdmin == false && session.User.Login != login {
		http.Error(w, "Users can only modify themselves", 403)
		return
	}

	profile, err := db.Profile(login)
	if err != nil || profile == nil {
		http.Error(w, "Failed to find user", 404)
		return
	}
	var cmd UpdateUser
	if err = json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Failed to bind command object. Are you sending JSON? "+err.Error(), 400)
		return
	}
	err = db.HideUser(cmd.Hide, profile.User.Login)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	profile.User.Hide = cmd.Hide
	session.User.Hide = cmd.Hide
	jsonResponse(w, 200, profile)
}

// Delete allows admins to easily expunge old data
func Delete(w http.ResponseWriter, r *http.Request) {
	session := sessions.GetEntry(r)
	if session.User.IsAdmin == false {
		http.Error(w, "Only admins can delete users", 403)
		return
	}

	login := r.PathValue("login")

	err := db.Delete(login)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	jsonResponse(w, 200, "deleted")
}

func jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
