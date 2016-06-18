package web

import (
	"net/http"
	"log"
	"encoding/json"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func topLangs(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := map[string]interface{}{}
		data["langs"] = cmd.PopularLanguages()
		if time, err := cmd.LastRun(); err == nil {
			data["lastrun"] = time
		}
		render(w, data)
	}
}

func topDevs(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := map[string]interface{}{}
		data["devs"] = cmd.PopularDevs()
		if time, err := cmd.LastRun(); err == nil {
			data["lastrun"] = time
		}
		render(w, data)
	}
}

func topOrgs(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data := map[string]interface{}{}
		data["devs"] = cmd.PopularOrgs()
		if time, err := cmd.LastRun(); err == nil {
			data["lastrun"] = time
		}
		render(w, data)
	}
}

func profile(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := map[string]interface{}{}
		profile, err := cmd.Profile(p.ByName("profile"))
		if err != nil {
			log.Println(err)
		}
		if profile != nil {
			data["profile"] = profile
		}
		render(w, data)
	}
}

func language(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		pageParam := r.URL.Query().Get("page")
		page := 0
		if pageParam != "" {
			var err error
			page, err = strconv.Atoi(pageParam)
			if err != nil {
				w.WriteHeader(400)
				return
			}
		}

		data := map[string]interface{}{}
		langs, userCount := cmd.Language(p.ByName("lang"), page)
		data["languages"] = langs
		data["count"] = userCount
		data["language"] = p.ByName("lang")
		data["page"] = page
		render(w, data)
	}
}

func search(ctx Context, cmd Commands) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := map[string]interface{}{}
		q := r.URL.Query().Get("q")
		kind := r.URL.Query().Get("type")
		if q != "" {
			results := cmd.Search(q, kind)
			if kind == "users" {
				data["results"] = results.([]User)
			} else if kind == "repos" {
				data["results"] = results.([]Repository)
			}
		}
		render(w, data)
	}
}

func render(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Error while rendering:", err)
		return
	}
}